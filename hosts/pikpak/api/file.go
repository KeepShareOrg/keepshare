// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/samber/lo"
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

func (t *fileTask) toFile(master, worker, link string) *model.File {
	if t == nil {
		return nil
	}

	status := t.Status
	if !comm.IsFinalStatus(t.Status) {
		status = comm.StatusRunning // make sure all not finished files are running
	}

	now := time.Now()
	return &model.File{
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
	}
}

// CreateFilesFromLink create files from link.
func (api *API) CreateFilesFromLink(ctx context.Context, master, worker, link string) (file *model.File, err error) {
	log.ContextWithFields(ctx, log.Fields{
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

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetBody(JSON{
			"kind":        "drive#file",
			"folder_type": "DOWNLOAD",
			"upload_type": "UPLOAD_TYPE_URL",
			"url":         JSON{"url": link},
		}).
		Post(apiURL("/drive/v1/files"))

	if err != nil {
		return nil, fmt.Errorf("create file err: %w", err)
	}

	log.WithContext(ctx).Debugf("create file response body: %s", body.Body())

	if err = e.Error(); err != nil {
		// TODO token expired
		return nil, fmt.Errorf("create file err: %w", err)
	}

	if r.Task.ID == "" {
		return nil, fmt.Errorf("create file got unexpected body: %s", body.Body())
	}

	file = r.Task.toFile(master, worker, link)

	// store file record.
	// file_id may be empty.
	if err = api.q.File.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(file); err != nil {
		return nil, err
	}

	return file, nil
}

// UpdateFilesStatus update files status.
// All files must belong to the same worker.
func (api *API) UpdateFilesStatus(ctx context.Context, workerUserID string, files []*model.File, updateRunningTasks ...bool) (err error) {
	token, err := api.getToken(ctx, workerUserID, false)
	if err != nil {
		return err
	}

	var e RespErr
	var r struct {
		//NextPageToken string `json:"next_page_token"`
		//ExpiresIn     int64  `json:"expires_in"`
		Tasks []*fileTask `json:"tasks"`
	}

	taskIDs := make([]string, 0, len(files))
	taskIDToFile := make(map[string]*model.File, len(files))
	for _, file := range files {
		taskIDs = append(taskIDs, file.TaskID)
		taskIDToFile[file.TaskID] = file
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
		return fmt.Errorf("query task err: %w", err)
	}

	log.WithContext(ctx).WithFields(map[string]any{
		"worker":  workerUserID,
		"task_id": strings.Join(taskIDs, ","),
	}).Debugf("query task resp body: %s", body.Body())

	if err = e.Error(); err != nil {
		// TODO token expired
		return fmt.Errorf("query task err: %w", err)
	}

	t := &api.q.File
	now := time.Now()
	for _, task := range r.Tasks {
		if task.Status == comm.StatusRunning && !lo.Contains(updateRunningTasks, true) {
			continue // do not update running task.
		}

		file := taskIDToFile[task.ID]
		file.FileID = task.FileID
		file.Status = task.Status
		file.IsDir = task.StatusSize > 1
		file.Size = int64(util.Atoi(task.FileSize))
		file.Name = task.FileName
		file.UpdatedAt = now

		if file.Status == comm.StatusOK {
			if callbackFns, ok := eventListeners[hosts.FileComplete]; ok {
				for _, callbackFn := range callbackFns {
					callbackFn(file.WorkerUserID, file.OriginalLinkHash)
				}
			}
		}

		t.Select(t.FileID, t.Status, t.IsDir, t.Size, t.Name, t.UpdatedAt).Where(t.TaskID.Eq(task.ID)).Updates(file)
	}

	// if query task is null, remvoe form pikpak_file table.
	if len(files) > len(r.Tasks) {
		shouldRemoveTaskIds := make([]string, 0)
		for _, f := range files[len(r.Tasks):] {
			if isTaskNotFound(r.Tasks, f.TaskID) {
				shouldRemoveTaskIds = append(shouldRemoveTaskIds, f.TaskID)
				continue
			}
		}
		log.WithContext(ctx).Debugf("shoule delete pikpak_file tasks: %v", shouldRemoveTaskIds)
		if len(shouldRemoveTaskIds) > 0 {
			t.Where(t.TaskID.In(shouldRemoveTaskIds...)).Delete()
		}
	}

	return nil
}

func isTaskNotFound(tasks []*fileTask, id string) bool {
	for _, t := range tasks {
		if t.ID == id {
			return false
		}
	}
	return true
}

func (api *API) handelTriggerChan() {
	for {
		// externalTriggerChan has higher priority than internalTriggerChan.
		select {
		case wfs := <-api.externalTriggerChan:
			api.updateRunningFiles(wfs.worker, wfs.files)
		default:
			select {
			case wfs := <-api.internalTriggerChan:
				api.updateRunningFiles(wfs.worker, wfs.files)
			default:
				time.Sleep(time.Second)
			}
		}
	}
}

func (api *API) recentTaskConsumer() {
	for {
		select {
		case wfs := <-api.recentTasksChan:
			api.updateRunningFiles(wfs.worker, wfs.files)
		default:
			time.Sleep(time.Second)
		}
	}
}

func (api *API) getRecentFilesFromDB() {
	createdAfter := time.Now().Add(-time.Minute * 10)
	for {
		workerFiles, token := api.getRecentFiles(createdAfter)
		if workerFiles == nil {
			createdAfter = token
			time.Sleep(time.Second)
		}
		if len(workerFiles) > 0 {
			for worker, files := range workerFiles {
				api.recentTasksChan <- runningFiles{worker: worker, files: files}
			}
		}
		if len(workerFiles) < 100 {
			time.Sleep(2 * time.Second)
		}
	}
}

func (api *API) getRecentFiles(createdAfter time.Time) (map[string][]*model.File, time.Time) {
	t := &api.q.File
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	files, err := t.WithContext(ctx).
		Where(
			t.Status.In(comm.StatusRunning, comm.StatusPending),
			t.CreatedAt.Gt(createdAfter),
		).
		Order(t.CreatedAt).
		Limit(comm.RunningFilesSelectLimit).
		Find()
	if err != nil {
		return nil, time.Now().Add(-time.Minute * 10)
	}

	m := map[string][]*model.File{}
	for _, f := range files {
		m[f.WorkerUserID] = append(m[f.WorkerUserID], f)
	}

	token := time.Now().Add(-time.Minute * 10)
	if len(files) > 0 {
		token = files[len(files)-1].CreatedAt
	}
	return m, token
}

func (api *API) triggerFilesFromDB() {
	getRunningFilesToken := &GetRunningFilesToken{
		UpdatedTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		OrderID:     0,
	}
	for {
		workerFiles, token := api.getRunningFiles(*getRunningFilesToken)
		if token != nil {
			getRunningFilesToken = token
		}
		log.Debugf("current internal trigger chan length: %v, from db length: %v", len(api.internalTriggerChan), len(workerFiles))
		if len(api.internalTriggerChan) < 10 && len(workerFiles) < 10 {
			getRunningFilesToken = &GetRunningFilesToken{
				UpdatedTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				OrderID:     0,
			}
		}
		if len(workerFiles) > 0 {
			for worker, files := range workerFiles {
				api.internalTriggerChan <- runningFiles{worker: worker, files: files}
			}
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func (api *API) updateRunningFiles(worker string, files []*model.File) {
	ctx := log.DataContext(context.Background(), log.DataContextOptions{
		Fields: log.Fields{"src": "scan_file"},
	})
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// update files status in batch, 100 per batch.
	batchSize := 100
	if len(files) > batchSize {
		for i := 0; i < len(files); i += batchSize {
			end := int(math.Min(float64(i+batchSize), float64(len(files))))
			if err := api.UpdateFilesStatus(ctx, worker, files[i:end]); err != nil {
				log.WithContext(ctx).WithField("worker", worker).WithError(err).Error("update files status err")
				return
			}
		}
	} else {
		if err := api.UpdateFilesStatus(ctx, worker, files); err != nil {
			log.WithContext(ctx).WithField("worker", worker).WithError(err).Error("update files status err")
			return
		}
	}

	// if all the files are completed, update immediately,
	// otherwise control the update frequency by redis.
	hasCompleted := false
	hasRunning := false
	for _, f := range files {
		if !comm.IsFinalStatus(f.Status) {
			hasRunning = true
		}
		if f.Status == comm.StatusOK {
			hasCompleted = true
		}
	}

	if !hasCompleted {
		// nothing changed, no need to update
		return
	}

	if hasRunning {
		key := fmt.Sprintf("pikpak:updateStorage:%s", worker)
		ok, err := api.Redis.SetNX(ctx, key, "", time.Minute).Result()
		if err == nil && !ok {
			// Updated within 1 minute, no need to update at this time
			return
		}
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	if err := api.UpdateWorkerStorage(ctx, worker); err != nil {
		log.WithContext(ctx).WithField("worker", worker).WithError(err).Error("update storage err")
	}
}

type GetRunningFilesToken struct {
	UpdatedTime time.Time
	OrderID     int64
}

// getRunningFiles returns a map of worker -> files.
func (api *API) getRunningFiles(token GetRunningFilesToken) (map[string][]*model.File, *GetRunningFilesToken) {
	t := &api.q.File
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status := t.Status.ColumnName().String()
	createdAt := t.CreatedAt.ColumnName().String()
	updatedAt := t.UpdatedAt.ColumnName().String()
	autoID := t.AutoID.ColumnName().String()

	w := fmt.Sprintf("`%s` in ('%s', '%s') AND (`%s`, `%s`) > ('%s', '%v') AND `%s`>'%s' AND TIMESTAMPDIFF(SECOND, `%s`, NOW())*60 > TIMESTAMPDIFF(SECOND, `%s`, `%s`)",
		status, comm.StatusRunning, comm.StatusPending,
		updatedAt, autoID, token.UpdatedTime.String(), token.OrderID,
		createdAt, time.Now().Add(-1*comm.RunningFilesMaxAge).Format(time.DateTime),
		updatedAt, createdAt, updatedAt,
	)

	var files []*model.File
	//err := api.Mysql.WithContext(gormutil.IgnoreTraceContext(ctx)).
	err := api.Mysql.WithContext(ctx).
		Where(w).
		Order(updatedAt).
		Order(t.AutoID).
		Limit(comm.RunningFilesSelectLimit).
		Find(&files).
		Error
	if err != nil && !gormutil.IsNotFoundError(err) {
		log.WithContext(ctx).Error("get running files err:", err)
		return nil, nil
	}

	if len(files) == 0 {
		return nil, nil
	}

	m := map[string][]*model.File{}
	for _, f := range files {
		m[f.WorkerUserID] = append(m[f.WorkerUserID], f)
	}

	var nextToken *GetRunningFilesToken = nil
	if len(files) > 0 {
		nextToken = &GetRunningFilesToken{
			UpdatedTime: files[len(files)-1].UpdatedAt,
			OrderID:     files[len(files)-1].AutoID,
		}
	}
	return m, nextToken
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

// TriggerRunningFile try to update the status of the running or pending files.
func (api *API) TriggerRunningFile(file *model.File) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	set, err := api.Redis.SetNX(ctx, "trigger_running:"+file.TaskID, "", 30*time.Second).Result()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("trigger running check err")
		return
	}
	if !set {
		log.WithContext(ctx).Debug("trigger running ignore task:", file.TaskID)
		return
	}

	select {
	case api.externalTriggerChan <- runningFiles{file.WorkerUserID, []*model.File{file}}:
	default:
		log.WithContext(ctx).Debug("externalTriggerChan maybe full")
	}
}
