// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gen"
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
	token, err := api.getToken(ctx, worker)
	if err != nil {
		return nil, err
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

	log.WithFields(map[string]any{
		"master": master,
		"worker": worker,
		"link":   link,
	}).Debugf("create file response body: %s", body.Body())

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
	if err = api.q.File.WithContext(ctx).Create(file); err != nil {
		return nil, err
	}

	return file, nil
}

// UpdateFilesStatus update files status.
// All files must belong to the same worker.
func (api *API) UpdateFilesStatus(ctx context.Context, workerUserID string, files []*model.File, updateRunningTasks ...bool) (err error) {
	token, err := api.getToken(ctx, workerUserID)
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

	log.WithFields(map[string]any{
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

		t.Select(t.FileID, t.Status, t.IsDir, t.Size, t.Name, t.UpdatedAt).Where(t.TaskID.Eq(task.ID)).Updates(file)
	}

	return nil
}

func (api *API) checkFilesBackground() {
	const interval = 1 * time.Second

	update := func(api *API, worker string, files []*model.File) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := api.UpdateFilesStatus(ctx, worker, files, true); err != nil {
			log.WithField("worker", worker).WithError(err).Error("update files status err")
			return
		}

		// if all the files are completed, update immediately,
		// otherwise control the update frequency by redis.
		hasCompleted := false
		hasRunning := false
		for _, f := range files {
			if !comm.IsFinalStatus(f.Status) {
				hasRunning = true
			} else if f.Status == comm.StatusOK {
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
			log.WithField("worker", worker).WithError(err).Error("update storage err")
		}
	}

	for {
		time.Sleep(interval)

		workerToFiles := api.getRunningFiles()
		for worker, files := range workerToFiles {
			update(api, worker, files)
		}

		if len(workerToFiles) == 0 {
			time.Sleep(5 * interval)
		}
	}
}

// getRunningFiles returns a map of worker -> files.
func (api *API) getRunningFiles() map[string][]*model.File {
	t := &api.q.File
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status := t.Status.ColumnName().String()
	createdAt := t.CreatedAt.ColumnName().String()
	updatedAt := t.UpdatedAt.ColumnName().String()

	w := fmt.Sprintf("`%s`='%s' AND `%s`>'%s' AND TIMESTAMPDIFF(SECOND, `%s`, NOW())*60 > TIMESTAMPDIFF(SECOND, `%s`, `%s`)",
		status, comm.StatusRunning,
		createdAt, time.Now().Add(-1*comm.RunningFilesMaxAge).Format(time.DateTime),
		updatedAt, createdAt, updatedAt,
	)

	var files []*model.File
	err := api.Mysql.WithContext(gormutil.IgnoreTraceContext(ctx)).
		Where(w).
		Order(updatedAt).
		Limit(comm.RunningFilesSelectLimit).
		Find(&files).
		Error
	if err != nil && !gormutil.IsNotFoundError(err) {
		log.Error("get running files err:", err)
		return nil
	}

	if len(files) == 0 {
		return nil
	}

	log.Debugf("condition: %s, running files count: %d", w, len(files))

	m := map[string][]*model.File{}
	for _, f := range files {
		m[f.WorkerUserID] = append(m[f.WorkerUserID], f)
	}
	return m
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
	token, err := api.getToken(ctx, worker)
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

	if log.IsLevelEnabled(log.DebugLevel) {
		log.WithFields(map[string]any{
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
