// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	q "github.com/KeepShareOrg/keepshare/pkg/queue"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	statisticTask      = "statistic"
	statisticTaskDelay = 2 * time.Minute
)

var queue *q.Client

type getStatisticsMessage struct {
	RecordID int64 `json:"record"`
}

func getStatisticsLater(recordID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := fmt.Sprintf("%s:%d", statisticTask, recordID)
	ok, _ := config.Redis().SetNX(ctx, key, "", statisticTaskDelay).Result()
	if ok {
		payload, _ := json.Marshal(getStatisticsMessage{RecordID: recordID})
		if t, err := queue.Enqueue(statisticTask, payload, asynq.ProcessIn(statisticTaskDelay)); err != nil {
			log.WithField(constant.Error, err).Errorf("enqueue statistics task for record %d err: %v", recordID, err)
		} else {
			log.Debugf("enqueue statistics task for record %d done, task id: %s", recordID, t.ID)
		}
	}
}

func handleGetStatistics(ctx context.Context, task *asynq.Task) error {
	var msg getStatisticsMessage
	_ = json.Unmarshal(task.Payload(), &msg)
	if msg.RecordID <= 0 {
		return nil // ignore invalid msg
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t := query.SharedLink
	rec, err := t.WithContext(ctx).Select(t.AutoID, t.UserID, t.Host, t.HostSharedLink, t.Stored).Where(t.AutoID.Eq(msg.RecordID)).Take()
	if err == gorm.ErrRecordNotFound {
		return nil // record deleted
	}
	if err != nil {
		return err // auto retry later
	}

	host := hosts.Get(rec.Host)
	if host == nil {
		return nil // host deleted
	}

	link := rec.HostSharedLink
	log := log.WithFields(Map{constant.SharedLink: link, constant.UserID: rec.UserID})

	stats, err := host.GetStatistics(ctx, rec.UserID, []string{link})
	if err != nil {
		log.WithField(constant.Error, err).Error("get statistics error")
		return err
	}

	stat, ok := stats[link]
	if !ok {
		return nil
	}

	now := time.Now()
	update := &model.SharedLink{AutoID: rec.AutoID}
	if rec.Stored < stat.Stored {
		update.LastStoredAt = now
		update.LastVisitedAt = now
	}
	update.Visitor = stat.Visitor
	update.Stored = stat.Stored
	update.Revenue = stat.Revenue
	update.UpdatedAt = now

	_, err = t.WithContext(ctx).Updates(update)
	if err != nil {
		log.WithField(constant.Error, err).Error("update statistics error")
		return err
	}

	log.Debugf("update statistics done: %+v", stat)
	return nil
}
