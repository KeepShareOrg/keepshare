package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/queue"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/hibiken/asynq"
	"math"
	"sync"
	"time"
)

type TaskManager struct {
	q         *query.Query
	d         *hosts.Dependencies
	api       *API
	taskQueue *queue.Queue
}

func newTaskManager(q *query.Query, api *API, d *hosts.Dependencies) *TaskManager {
	// init queue with redis, select another db = db + 2.
	redisOpt := *(config.Redis().Options())
	redisOpt.DB = (redisOpt.DB + 2) % 16
	queueIns := queue.New(redisOpt, map[string]int{
		constant.TaskQueuePikPak: 6,
	})

	m := &TaskManager{
		q:         q,
		d:         d,
		api:       api,
		taskQueue: queueIns,
	}

	// register handler
	if err := queueIns.Client().RegisterHandler(constant.TaskTypePPTask, asynq.HandlerFunc(m.ppTasksHandler)); err != nil {
		log.Errorf("register running task handler err: %v", err)
		return nil
	}

	queueIns.Run()

	return m
}

var taskManager *TaskManager

func GetTaskManagerInstance(q *query.Query, api *API, d *hosts.Dependencies) *TaskManager {
	if taskManager == nil {
		if d == nil {
			return nil
		}
		taskManager = newTaskManager(q, api, d)
	}

	return taskManager
}

var allTaskEnqueue func() = nil

func (t *TaskManager) Start(ctx context.Context) {
	for {
		statusTypes := []string{comm.StatusPending, comm.StatusRunning}
		if allTaskEnqueue == nil {
			allTaskEnqueue = sync.OnceFunc(func() {
				err := t.EnQueueAllTasksFromDB(ctx)
				if err != nil {
					log.Errorf("EnQueueAllTasksFromDB failed: %v", err)
				}
			})
		}
		allTaskEnqueue()

		if ppTasks, err := t.GetRecentTasksFromDB(ctx, statusTypes, -10*time.Minute, 3000); err == nil {
			t.taskEnQueue(constant.TaskQueuePikPak, constant.TaskTypePPTask, ppTasks)
		}

		time.Sleep(time.Second)
	}
}

func (t *TaskManager) EnQueueAllTasksFromDB(ctx context.Context) error {
	statusTypes := []string{comm.StatusPending, comm.StatusRunning}

	currentAutoId := int64(0)
	for {
		f := t.q.File
		tasks, err := f.WithContext(gormutil.IgnoreTraceContext(ctx)).
			Where(
				f.Status.In(statusTypes...),
				f.AutoID.Gt(currentAutoId),
			).
			Order(f.AutoID).
			Limit(2000).
			Find()
		if err != nil {
			log.Errorf("GetRecentRunningTasksFromDB failed: %v", err)
			return err
		}
		if len(tasks) == 0 {
			break
		}

		t.taskEnQueue(constant.TaskQueuePikPak, constant.TaskTypePPTask, tasks)
		currentAutoId = tasks[len(tasks)-1].AutoID
	}

	return nil
}

func (t *TaskManager) GetRecentTasksFromDB(ctx context.Context, status []string, beforeTime time.Duration, limit int) ([]*model.File, error) {
	f := t.q.File
	tasks, err := f.WithContext(ctx).
		Where(
			f.Status.In(status...),
			f.CreatedAt.Gte(time.Now().Add(beforeTime)),
		).
		Order(f.CreatedAt).
		Limit(limit).
		Find()
	if err != nil {
		log.Errorf("GetRecentRunningTasksFromDB failed: %v", err)
		return nil, err
	}

	return tasks, nil
}

type PikPakTask struct {
	AutoID           int64
	MasterUserID     string
	WorkerUserID     string
	FileID           string
	TaskID           string
	Status           string
	IsDir            bool
	Size             int64
	Name             string
	OriginalLinkHash string
}

func (t *TaskManager) taskEnQueue(queueType, taskType string, tasks []*model.File) {
	for _, task := range tasks {
		payload, err := json.Marshal(&PikPakTask{
			AutoID:           task.AutoID,
			MasterUserID:     task.MasterUserID,
			WorkerUserID:     task.WorkerUserID,
			FileID:           task.FileID,
			TaskID:           task.TaskID,
			Status:           task.Status,
			IsDir:            task.IsDir,
			Size:             task.Size,
			Name:             task.Name,
			OriginalLinkHash: task.OriginalLinkHash,
		})
		if err != nil {
			log.Errorf("payload marshal failed: %v", err)
			continue
		}

		_, err = t.taskQueue.Client().Enqueue(
			taskType,
			payload,
			asynq.Unique(48*time.Hour),
			asynq.MaxRetry(math.MaxInt32),
			asynq.Queue(queueType),
		)
		if err != nil && !errors.Is(err, asynq.ErrDuplicateTask) {
			log.Errorf("taskEnQueue failed: %v", err)
		}
	}
}

// ppTasksHandler if return nil, the task will remove from queue
func (t *TaskManager) ppTasksHandler(ctx context.Context, task *asynq.Task) error {
	var ppTask PikPakTask
	err := json.Unmarshal(task.Payload(), &ppTask)
	if err != nil {
		log.Errorf("unmarshal task payload failed: %v", err)
		return nil
	}

	status, err := t.api.queryTaskStatus(ctx, ppTask.WorkerUserID, ppTask.TaskID)
	if err != nil {
		log.Errorf("%#v query task status failed: %v", ppTask, err)
		return err
	}
	log.Debugf("file id: %v, status: %v", status.FileID, status.Status)

	file := &model.File{
		AutoID:           ppTask.AutoID,
		MasterUserID:     ppTask.MasterUserID,
		WorkerUserID:     ppTask.WorkerUserID,
		FileID:           status.FileID,
		TaskID:           ppTask.TaskID,
		Status:           ppTask.Status,
		IsDir:            ppTask.IsDir,
		Size:             ppTask.Size,
		Name:             status.FileName,
		OriginalLinkHash: ppTask.OriginalLinkHash,
		UniqueHash:       fmt.Sprintf("%s:%s", ppTask.WorkerUserID, lk.Hash(ppTask.OriginalLinkHash)),
	}

	switch status.Status {
	case comm.StatusRunning:
		//if the task is running but progress is over 90% that means the task is complete,
		//so we should update the status can create shared link
		if status.Progress >= 90 && status.FileID != "" {
			file.Status = comm.StatusOK
			err := t.handleStatusOKTask(ctx, file)
			if err != nil {
				log.Errorf("handle status PHASE_TYPE_RUNNING but progress over 90% task error: ", err)
				return err
			}
		}
		return nil
	case comm.StatusOK:
		err := t.handleStatusOKTask(ctx, file)
		if err != nil {
			log.Errorf("handle status PHASE_TYPE_COMPLETE task error: ", err)
			return err
		}
		log.Debugf("file complete: %v", file.OriginalLinkHash)
		if callbackFns, ok := eventListeners[hosts.FileComplete]; ok {
			for _, callbackFn := range callbackFns {
				callbackFn(file.WorkerUserID, file.OriginalLinkHash)
			}
		}
		return nil
	case comm.StatusPending:
		err := t.handleStatusPendingTask(ctx, file)
		if err != nil {
			log.Errorf("handle status PHASE_TYPE_PENDING task error: ", err)
			return err
		}
		return nil
	case comm.StatusError:
		err := t.handleStatusErrorTask(ctx, file)
		if err != nil {
			log.Errorf("handle status PHASE_TYPE_ERROR task error: ", err)
			return err
		}
		return nil
	default:
		return fmt.Errorf("task status not complete: %v", status.Status)
	}
}

// handleStatusOKTask handle status PHASE_TYPE_COMPLETE
func (t *TaskManager) handleStatusOKTask(ctx context.Context, file *model.File) error {
	f := &t.q.File
	_, err := f.WithContext(ctx).Where(f.AutoID.Eq(file.AutoID)).Updates(&model.File{
		Status:    comm.StatusOK,
		FileID:    file.FileID,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return err
	}
	return nil
}

// handleStatusPendingTask handle status PHASE_TYPE_PENDING
func (t *TaskManager) handleStatusPendingTask(ctx context.Context, file *model.File) error {
	f := &t.q.File
	_, err := f.WithContext(ctx).Where(f.AutoID.Eq(file.AutoID)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// handleStatusErrorTask handle status PHASE_TYPE_ERROR
func (t *TaskManager) handleStatusErrorTask(ctx context.Context, file *model.File) error {
	f := &t.q.File
	_, err := f.WithContext(ctx).Where(f.AutoID.Eq(file.AutoID)).Update(f.Status, comm.StatusError)
	if err != nil {
		return err
	}
	return nil
}
