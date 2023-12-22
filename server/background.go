// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/spf13/viper"

	"github.com/KeepShareOrg/keepshare/hosts"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
)

type asyncBackgroundTask struct {
	concurrency     int
	unCompletedChan chan int64
}

func (a *asyncBackgroundTask) pushAsyncTask(linkID int64) {
	a.unCompletedChan <- linkID
}

func (a *asyncBackgroundTask) getTaskFromDB() {
	getUncompletedToken := &getUnCompletedToken{
		UpdatedTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		OrderID:     0,
	}
	for {
		unCompleteTasks, token, err := getUnCompletedSharedLinks(1000, *getUncompletedToken)
		if token != nil {
			getUncompletedToken = token
		}
		if err != nil {
			log.Errorf("get uncompleted tasks err: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Infof("current uncompleted chan length: %v, from db length: %v", len(a.unCompletedChan), len(unCompleteTasks))
		if len(a.unCompletedChan) < 10 && len(unCompleteTasks) < 10 {
			getUncompletedToken = &getUnCompletedToken{
				UpdatedTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				OrderID:     0,
			}
		}
		// this file in `pikpak_file` status is `PHASE_TYPE_COMPLETE`
		if len(unCompleteTasks) > 0 {
			for _, task := range unCompleteTasks {
				a.pushAsyncTask(task.AutoID)
			}
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func (a *asyncBackgroundTask) batchProcessCompleteTask() {
	for {
		var sharedLinks []*model.SharedLink
		// this file in `pikpak_file` status is `PHASE_TYPE_COMPLETE`
		r := fmt.Sprintf(
			"select * from `keepshare_shared_link` where state = '%s' and original_link_hash in (select original_link_hash from `pikpak_file` where status = '%s' AND updated_at > '%s')",
			share.StatusCreated.String(),
			comm.StatusOK,
			time.Now().Add(-24*time.Hour).Format(time.DateTime),
		)
		if err := config.MySQL().Raw(r).Scan(&sharedLinks).Error; err != nil {
			log.Errorf("select shared link err: %v", err)
		}

		if len(sharedLinks) > 0 {
			log.Debugf("handle complete task length: %v", len(sharedLinks))
			ctx := context.Background()
			redis := config.Redis()
			for _, task := range sharedLinks {
				key := fmt.Sprintf("%v", task.AutoID)
				set := redis.SetNX(ctx, key, "1", 10*time.Minute).Val()
				if set {
					log.Debugf("handle complete task: %#v", task)
					a.pushAsyncTask(task.AutoID)
				}
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func (a *asyncBackgroundTask) startConsumer() {
	for {
		select {
		case id := <-a.unCompletedChan:
			a.taskConsumer(id)
		default:
			time.Sleep(time.Second)
		}
	}
}

func (a *asyncBackgroundTask) taskConsumer(linkID int64) {
	// get latest record
	task, err := query.SharedLink.WithContext(context.Background()).Where(query.SharedLink.AutoID.Eq(linkID)).Take()
	if err != nil {
		return
	}
	// return if the task is completed.
	if task.State == share.StatusOK.String() {
		return
	}

	ctx := log.DataContext(context.Background(), log.DataContextOptions{
		Fields: log.Fields{
			"src":           "scan_share_record",
			constant.UserID: task.UserID,
		},
	})
	lg := log.WithContext(ctx)

	host := hosts.Get(task.Host)
	if host == nil {
		lg.Errorf("host not found: %s", task.Host)
		return
	}

	lg.Debugf("handle uncomplete task: %#v", task)
	sharedLinks, err := host.CreateFromLinks(
		ctx,
		task.UserID,
		[]string{task.OriginalLink},
		task.CreatedBy,
	)
	if err != nil {
		lg.Errorf("create share link error: %v", err.Error())
		update := model.SharedLink{
			UpdatedAt: time.Now(),
			State:     share.StatusError.String(),
			Error:     err.Error(),
		}
		if gormutil.IsNotFoundError(err) {
			lg.Debugf("create share link not found")
			//update.CreatedAt = time.Now()
		}
		if _, err = query.SharedLink.
			WithContext(ctx).
			Where(query.SharedLink.AutoID.Eq(task.AutoID)).
			Updates(update); err != nil {
			lg.Errorf("update share link updated_at error: %v", err.Error())
		}
		return
	}

	sh := sharedLinks[task.OriginalLink]
	if sh == nil {
		lg.Errorf("link not found: %s", task.OriginalLink)
		if _, err = query.SharedLink.
			WithContext(ctx).
			Where(query.SharedLink.AutoID.Eq(task.AutoID)).
			Updates(model.SharedLink{
				//CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				State:     share.StatusPending.String(),
			}); err != nil {
			lg.Errorf("update share link updated_at error: %v", err.Error())
		}
		return
	}

	// if task processing duration grate than 48 hour, it's failed
	if sh.State == share.StatusCreated && time.Now().Sub(sh.CreatedAt).Hours() > 48 {
		if _, err := query.SharedLink.
			WithContext(ctx).
			Where(query.SharedLink.AutoID.Eq(task.AutoID)).
			Updates(model.SharedLink{
				UpdatedAt: time.Now(),
				State:     share.StatusError.String(),
			}); err != nil {
			lg.Errorf("update share link error: %v", err.Error())
		}
		return
	}

	if sh.State == share.StatusOK || sh.State == share.StatusCreated {
		now := time.Now()
		update := &model.SharedLink{
			State:              sh.State.String(),
			UpdatedAt:          now,
			Size:               sh.Size,
			Visitor:            sh.Visitor,
			Stored:             sh.Stored,
			Revenue:            sh.Revenue,
			Title:              sh.Title,
			HostSharedLinkHash: lk.Hash(sh.HostSharedLink),
			HostSharedLink:     sh.HostSharedLink,
		}

		if _, err = query.SharedLink.
			WithContext(ctx).
			Where(query.SharedLink.AutoID.Eq(task.AutoID)).
			Updates(update); err != nil {
			lg.Errorf("update share link state error: %v", err.Error())
		}
		return
	}

	if _, err = query.SharedLink.
		WithContext(ctx).
		Where(query.SharedLink.AutoID.Eq(task.AutoID)).
		Update(query.SharedLink.UpdatedAt, time.Now()); err != nil {
		lg.Errorf("update share link updated_at error: %v", err.Error())
	}
}

func (a *asyncBackgroundTask) run() {
	go a.getTaskFromDB()
	go a.batchProcessCompleteTask()

	for _, h := range hosts.GetAll() {
		h.AddEventListener(hosts.FileComplete, func(userID, originalLinkHash string) {
			link, err := query.SharedLink.WithContext(context.Background()).Where(query.SharedLink.OriginalLinkHash.Eq(originalLinkHash)).Take()
			if err != nil {
				log.Errorf("get shared link err: %v", err)
				return
			}
			log.Debugf("get shared link from listener callback: %#v", link)
			a.taskConsumer(link.AutoID)
		})
	}

	wg := sync.WaitGroup{}
	for i := 0; i < a.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.startConsumer()
		}()
	}
	wg.Wait()
}

func newAsyncBackgroundTask(concurrency int) *asyncBackgroundTask {
	chSize := viper.GetInt("background_task_channel_size")
	if chSize <= 0 {
		chSize = 16 * 1024
	}

	return &asyncBackgroundTask{
		concurrency:     concurrency,
		unCompletedChan: make(chan int64, chSize),
	}
}

var abt *asyncBackgroundTask

func getAsyncBackgroundTaskInstance() *asyncBackgroundTask {
	if abt == nil {
		concurrency := viper.GetInt("background_task_concurrency")
		if concurrency <= 0 {
			concurrency = 16
		}
		abt = newAsyncBackgroundTask(concurrency)
	}
	return abt
}

type getUnCompletedToken struct {
	UpdatedTime time.Time
	OrderID     int64
}

// getUnCompletedSharedLinks get shared links that status in pending or created
func getUnCompletedSharedLinks(limitSize int, token getUnCompletedToken) ([]*model.SharedLink, *getUnCompletedToken, error) {
	s := query.SharedLink
	state := s.State.ColumnName().String()
	createdAt := s.CreatedAt.ColumnName().String()
	updatedAt := s.UpdatedAt.ColumnName().String()
	autoID := s.AutoID.ColumnName().String()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	unCompleteTasks := make([]*model.SharedLink, 0)

	w := fmt.Sprintf("`%s` in ('%s', '%s') AND (`%s`, `%s`) > ('%s', '%v') AND TIMESTAMPDIFF(SECOND, `%s`, NOW())*300 > TIMESTAMPDIFF(SECOND, `%s`, `%s`)",
		state, string(share.StatusPending), string(share.StatusCreated),
		updatedAt, autoID, token.UpdatedTime.String(), token.OrderID,
		updatedAt, createdAt, updatedAt,
	)
	//err := config.MySQL().WithContext(gormutil.IgnoreTraceContext(ctx)).
	err := config.MySQL().WithContext(ctx).
		Where(w).
		Order(fmt.Sprintf("%v DESC", state)).
		Order(updatedAt).
		Order(autoID).
		Limit(limitSize).
		Find(&unCompleteTasks).
		Error

	if err != nil {
		return nil, nil, err
	}

	var nextToken *getUnCompletedToken = nil
	if len(unCompleteTasks) > 0 {
		nextToken = &getUnCompletedToken{
			UpdatedTime: unCompleteTasks[len(unCompleteTasks)-1].UpdatedAt,
			OrderID:     unCompleteTasks[len(unCompleteTasks)-1].AutoID,
		}
	}
	return unCompleteTasks, nextToken, nil
}
