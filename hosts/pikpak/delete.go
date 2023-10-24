// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gorm.io/gorm/clause"
)

const (
	statusToDo             = "TODO"
	taskTypeSyncWorkerInfo = "pikpak_sync_worker_info"
)

// Delete delete shared links by original links.
func (p *PikPak) Delete(ctx context.Context, userID string, originalLinks []string) error {
	if len(originalLinks) == 0 {
		return nil
	}

	// get master account
	master, err := p.q.MasterAccount.WithContext(ctx).
		Select(p.q.MasterAccount.UserID).
		Where(p.q.MasterAccount.KeepshareUserID.Eq(userID)).
		Take()
	if gormutil.IsNotFoundError(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query master err: %w", err)
	}

	hashes := make([]string, 0, len(originalLinks))
	for _, l := range originalLinks {
		l = strings.TrimSpace(l)
		if l != "" {
			hashes = append(hashes, lk.Hash(l))
		}
	}
	if len(hashes) == 0 {
		return nil
	}

	files, err := p.q.File.WithContext(ctx).Where(
		p.q.File.MasterUserID.Eq(master.UserID),
		p.q.File.OriginalLinkHash.In(hashes...),
	).Find()
	if err != nil && !gormutil.IsNotFoundError(err) {
		return fmt.Errorf("query files err: %w", err)
	}
	if len(files) == 0 {
		return nil
	}

	ds := make([]*model.DeleteQueue, 0, len(files))
	now := time.Now()
	for _, f := range files {
		ds = append(ds, &model.DeleteQueue{
			WorkerUserID:     f.WorkerUserID,
			OriginalLinkHash: f.OriginalLinkHash,
			Status:           statusToDo,
			CreatedAt:        now,
			NextTrigger:      now,
			Ext:              "",
		})
	}

	if err := p.q.DeleteQueue.WithContext(ctx).Clauses(clause.Insert{Modifier: "IGNORE"}).Create(ds...); err != nil {
		return fmt.Errorf("insert to delete queue err: %w", err)
	}
	return nil
}

func (p *PikPak) deleteFilesBackground() {
	const interval = 1 * time.Second

	for {
		tasks, err := p.getToDoTasks()
		if err != nil && !gormutil.IsNotFoundError(err) {
			log.WithError(err).Errorf("get task from delete queue error")
			time.Sleep(5 * interval)
			continue
		}
		if len(tasks) == 0 {
			time.Sleep(interval)
			continue
		}

		workers := map[string]struct{}{}
		for _, task := range tasks {
			workers[task.WorkerUserID] = struct{}{}
			p.processDeleteTask(task)
		}
	}
}

func (p *PikPak) getToDoTasks() ([]*model.DeleteQueue, error) {
	const limit = 100

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t := &p.q.DeleteQueue
	return t.WithContext(gormutil.IgnoreTraceContext(ctx)).
		Where(t.Status.Eq(statusToDo)).
		Order(t.NextTrigger).
		Limit(limit).
		Find()
}

func (p *PikPak) processDeleteTask(task *model.DeleteQueue) {
	l := log.WithField("delete_queue", task)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	file, err := p.api.GetFileByOriginalLinkHash(ctx, "", task.WorkerUserID, task.OriginalLinkHash)
	if err != nil && !gormutil.IsNotFoundError(err) {
		log.WithError(err).Error("get file by original link hash err")
		p.addRetryTimesForDeleteTask(ctx, task)
		return
	}

	if file != nil {
		if err := p.api.DeleteShareByFileIDs(ctx, task.WorkerUserID, []string{file.FileID}); err != nil {
			l.WithError(err).Errorf("delete share by file id %s err", file.FileID)
			p.addRetryTimesForDeleteTask(ctx, task)
			return
		}

		if err := p.api.DeleteFilesByIDs(ctx, task.WorkerUserID, []string{file.FileID}); err != nil {
			l.WithError(err).Errorf("delete file by file id %s err", file.FileID)
			p.addRetryTimesForDeleteTask(ctx, task)
			return
		}
	}

	_, _ = p.q.DeleteQueue.WithContext(ctx).Delete(task)

	bs, _ := json.Marshal(map[string]string{"worker": task.WorkerUserID})
	if info, err := p.Queue.Enqueue(taskTypeSyncWorkerInfo, bs, asynq.ProcessIn(30*time.Second)); err != nil {
		log.Errorf("enqueue task type: %s, payload: %s, err: %v", taskTypeSyncWorkerInfo, bs, err)
	} else {
		log.Debugf("enqueue task type: %s, payload: %s, response id: %s", taskTypeSyncWorkerInfo, bs, info.ID)
	}

	return
}

func (p *PikPak) addRetryTimesForDeleteTask(ctx context.Context, row *model.DeleteQueue) {
	const (
		keyTryTimes = "tryTimes"
		maxTryTimes = 3
		nextTrigger = 60 * time.Second
	)

	t := &p.q.DeleteQueue

	ext := make(map[string]any)
	_ = json.Unmarshal([]byte(row.Ext), &ext)
	tryTimes := cast.ToInt(ext[keyTryTimes])
	tryTimes++

	if tryTimes > maxTryTimes {
		_, _ = t.WithContext(context.Background()).Delete(row)
		return
	}

	ext[keyTryTimes] = tryTimes
	row.Ext = util.ToJSON(ext)
	row.NextTrigger = time.Now().Add(nextTrigger * time.Duration(tryTimes))

	_, _ = t.WithContext(ctx).Updates(row)
	return
}

func (p *PikPak) syncWorkerHandler(ctx context.Context, task *asynq.Task) error {
	m := make(map[string]string)
	_ = json.Unmarshal(task.Payload(), &m)
	worker := m["worker"]
	if worker == "" {
		return nil
	}

	return p.api.UpdateWorkerStorage(ctx, worker)
}
