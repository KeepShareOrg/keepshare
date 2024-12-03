// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

type fileTask struct {
	ID          string `json:"id"`
	CreatedTime string `json:"created_time"`
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	FileSize    string `json:"file_size"`
	Message     string `json:"message"`
	Status      string `json:"phase"`
	StatusSize  int    `json:"status_size"`
	Progress    int64  `json:"progress"`
	Params      struct {
		PredictType string `json:"predict_type"`
	} `json:"params"`
}

var eventListeners = map[hosts.EventType][]hosts.ListenerCallback{}

// AddEventListener add event listener
func (api *API) AddEventListener(event hosts.EventType, fn hosts.ListenerCallback) {
	eventListeners[event] = append(eventListeners[event], fn)
}

func (t *fileTask) toFile(keepshareUserID, master, worker, link string) *model.File {
	if t == nil {
		return nil
	}

	status := t.Status
	if !comm.IsFinalStatus(t.Status) {
		status = comm.StatusRunning // make sure all not finished files are running
	}

	now := time.Now()
	f := &model.File{
		MasterUserID:     master,
		WorkerUserID:     worker,
		FileID:           t.FileID,
		TaskID:           t.ID,
		Status:           status,
		IsDir:            t.StatusSize > 1,
		Size:             int64(util.Atoi(t.FileSize)),
		Name:             t.FileName,
		CreatedAt:        now,
		UpdatedAt:        now,
		OriginalLinkHash: lk.Hash(link),
		UniqueHash:       fmt.Sprintf("%s:%s", keepshareUserID, lk.Hash(link)),
	}
	if t.Status == comm.StatusError {
		f.Error = t.Message
	}

	return f
}

// CreateFilesFromLink create files from link.
func (api *API) CreateFilesFromLink(ctx context.Context, master, worker, link string) (file *model.File, err error) {
	log := log.WithContext(ctx).WithFields(log.Fields{
		"master": master,
		"worker": worker,
		"link":   link,
	})
	defer func() {
		if err != nil {
			log.WithContext(ctx).Error("CreateFilesFromLink err:", err)
		}
	}()

	token, err := api.getToken(ctx, worker, false)
	if err != nil {
		return nil, fmt.Errorf("get token err: %w", err)
	}

	var e RespErr
	var r struct {
		Task fileTask `json:"task"`
	}

	requestID := uuid.New().String()
	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetHeader("X-Request-Id", requestID).
		SetResult(&r).
		SetBody(JSON{
			"kind":        "drive#file",
			"folder_type": "DOWNLOAD",
			"upload_type": "UPLOAD_TYPE_URL",
			"url":         JSON{"url": link},
		}).
		Post(apiURL("/drive/v1/files"))

	if err != nil {
		return nil, fmt.Errorf("create file err: %w, request-id: %v", err, requestID)
	}

	log.WithContext(ctx).Debugf("create file response body: %s", body.Body())

	if err = e.Error(); err != nil {
		// TODO token expired
		return nil, fmt.Errorf("create file err: %w", err)
	}

	if r.Task.ID == "" {
		return nil, fmt.Errorf("create file got unexpected body: %s", body.Body())
	}

	ksUserID := ""
	ma := &api.q.MasterAccount
	if err := ma.WithContext(ctx).Where(ma.UserID.Eq(master)).Pluck(ma.KeepshareUserID, &ksUserID); err != nil {
		return nil, err
	}

	file = r.Task.toFile(ksUserID, master, worker, link)

	// store file record.
	// file_id may be empty.
	if err = api.q.File.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(file); err != nil {
		return nil, err
	}

	return file, nil
}

// queryTasksStatus query tasks status
func (api *API) queryTasksStatus(ctx context.Context, workerUserID string, taskIDs []string) (fileTasks []*fileTask, queryTasksErr error) {
	token, err := api.getToken(ctx, workerUserID, false)
	if err != nil {
		return nil, err
	}

	var e RespErr
	var r struct {
		Tasks []*fileTask `json:"tasks"`
	}

	filters := map[string]any{
		"id": map[string]any{
			"in": strings.Join(taskIDs, ","),
		},
	}
	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetQueryParams(map[string]string{
			"type":    "offline",
			"limit":   "10000",
			"filters": util.ToJSON(filters),
		}).
		Get(apiURL("/drive/v1/tasks"))

	if err != nil {
		return nil, fmt.Errorf("query task err: %w", err)
	}

	log.WithContext(ctx).WithFields(map[string]any{
		"worker":  workerUserID,
		"task_id": strings.Join(taskIDs, ","),
	}).Debugf("query task resp body: %s", body.Body())

	if err = e.Error(); err != nil {
		// token expired, need to re-login this account
		if e.ErrorCode == 16 {
			go func() {
				//signIn
				w := &api.q.WorkerAccount
				worker, err := w.Where(w.UserID.Eq(workerUserID)).Take()
				if err != nil {
					return
				}
				if _, err := api.signIn(context.Background(), worker.UserID, worker.Email, worker.Password); err == nil {
					log.WithFields(map[string]any{
						"worker": worker.UserID,
					}).Debugf("re-sign in success.")
					tasks, err := api.queryTasksStatus(ctx, workerUserID, taskIDs)
					if err != nil {
						queryTasksErr = err
						return
					}
					fileTasks = tasks
				}
			}()
		}
		return nil, fmt.Errorf("query task err: %w", err)
	}

	return r.Tasks, nil
}

// queryTaskStatus query single task status
func (api *API) queryTaskStatus(ctx context.Context, workerUserID string, taskID string) (*fileTask, error) {
	tasks, err := api.queryTasksStatus(ctx, workerUserID, []string{taskID})
	if err != nil {
		return nil, err
	}
	if len(tasks) < 1 {
		return nil, fmt.Errorf("task not found")
	}
	task := tasks[0]

	// if task progress is over 95%, it can create share link
	if task.Progress > 95 && task.Progress < 100 {
		task.Progress = 100
		percent, err := api.QuerySubTasksCompleteSizePercent(ctx, workerUserID, task.ID)
		if err == nil && percent > 0.95 {
			log.Infof("should be create share link, because percent is over 95%.", percent)
			task.Status = comm.StatusOK
		}
	}

	if task.Status == comm.StatusOK {
		key := fmt.Sprintf("pikpak:updateStorage:%s", workerUserID)
		ok, err := api.Redis.SetNX(ctx, key, "", time.Minute).Result()
		if err == nil && !ok {
			// Updated within 1 minute, no need to update at this time
			return task, nil
		}

		go func() {
			ctx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel2()
			if err := api.UpdateWorkerStorage(ctx, workerUserID); err != nil {
				log.WithContext(ctx).WithField("worker", workerUserID).WithError(err).Error("update storage err")
			}
		}()
	}
	return task, nil
}

// QuerySubTasksCompleteSizePercent query sub-tasks complete size percent.
func (api *API) QuerySubTasksCompleteSizePercent(ctx context.Context, worker, taskID string) (float64, error) {
	token, err := api.getToken(ctx, worker, false)
	if err != nil {
		return 0, fmt.Errorf("get token err: %w", err)
	}

	var e RespErr
	var r struct {
		Statuses []struct {
			FileSize string `json:"file_size"`
			Phase    string `json:"phase"`
		} `json:"statuses"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetQueryParams(map[string]string{
			"limit": "100",
		}).
		Get(apiURL(fmt.Sprintf("/drive/v1/task/%s/statuses", taskID)))

	if err != nil {
		return 0, fmt.Errorf("query sub-task status err: %w", err)
	}

	log.WithContext(ctx).Debugf("query sub-task status resp body: %s", body.Body())

	if err = e.Error(); err != nil {
		return 0, err
	}

	totalSize, completeSize := 0, 0
	for _, status := range r.Statuses {
		totalSize += util.Atoi(status.FileSize)
		if status.Phase == comm.StatusOK {
			completeSize += util.Atoi(status.FileSize)
		}
	}

	if totalSize == 0 {
		return 0, nil
	}

	percent := float64(completeSize) / float64(totalSize)
	return percent, nil
}

// GetFileByOriginalLinkHash get file by original link.
// only one of master or worker is required.
func (api *API) GetFileByOriginalLinkHash(ctx context.Context, master string, worker string, originalLinkHash string) (*model.File, error) {
	t := &api.q.File
	var w []gen.Condition
	if master != "" {
		w = append(w, t.MasterUserID.Eq(master))
	} else if worker != "" {
		w = append(w, t.WorkerUserID.Eq(worker))
	}
	if len(w) == 0 {
		return nil, fmt.Errorf("one of master or worker is required")
	}

	w = append(w, t.OriginalLinkHash.Eq(originalLinkHash))
	return t.WithContext(ctx).Where(w...).Take()
}

func (api *API) DeleteFilesByIDs(ctx context.Context, worker string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return nil
	}
	token, err := api.getToken(ctx, worker, false)
	if err != nil {
		return err
	}

	var e RespErr
	var r struct {
		TaskID string `json:"task_id"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetBody(JSON{"ids": fileIDs}).
		Post(apiURL("/drive/v1/files:batchDelete"))

	if err != nil {
		return fmt.Errorf("delete files err: %w", err)
	}

	if log.IsDebugEnabled() {
		log.WithContext(ctx).WithFields(map[string]any{
			"worker":   worker,
			"file_ids": fileIDs,
		}).Debugf("delete files response body: %s", body.Body())
	}

	if err = e.Error(); err != nil {
		// TODO token expired
		return fmt.Errorf("delete files err: %w", err)
	}

	// delete file records.
	if _, err = api.q.File.WithContext(ctx).Where(api.q.File.FileID.In(fileIDs...)).Delete(); err != nil {
		return fmt.Errorf("delete files from db err: %w", err)
	}

	return nil
}
