// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	log "github.com/sirupsen/logrus"
)

var checkBatchSize = 100

// asyncTaskCheckBackground checks uncompleted shared links task in background
func asyncTaskCheckBackground() {
	for f := 0; ; f++ {
		if f != 0 {
			time.Sleep(1 * time.Second)
		}

		s := query.SharedLink
		state := s.State.ColumnName().String()
		createdAt := s.CreatedAt.ColumnName().String()
		updatedAt := s.UpdatedAt.ColumnName().String()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		unCompleteTasks := make([]*model.SharedLink, 0)

		w := fmt.Sprintf("`%s` in ('%s', '%s') AND `%s`>'%s' AND TIMESTAMPDIFF(SECOND, `%s`, NOW())*60 > TIMESTAMPDIFF(SECOND, `%s`, `%s`)",
			state, string(share.StatusPending), string(share.StatusCreated),
			createdAt, time.Now().Add(-1*comm.RunningFilesMaxAge).Format(time.DateTime),
			updatedAt, createdAt, updatedAt,
		)
		err := config.MySQL().WithContext(gormutil.IgnoreTraceContext(ctx)).
			Where(w).
			Order(updatedAt).
			Limit(checkBatchSize).
			Find(&unCompleteTasks).
			Error

		if err != nil {
			log.Errorf("query uncomplete share link task error: %v", err.Error())
			continue
		}

		completeTasks := make([]*model.SharedLink, 0)
		failedTasks := make([]*model.SharedLink, 0)
		for _, unCompleteTask := range unCompleteTasks {
			host := hosts.Get(unCompleteTask.Host)
			if host == nil {
				log.Errorf("host not found: %s", unCompleteTask.Host)
				continue
			}

			sharedLinks, err := host.CreateFromLinks(
				context.Background(),
				unCompleteTask.UserID,
				[]string{unCompleteTask.OriginalLink},
				unCompleteTask.CreatedBy,
			)
			if err != nil {
				log.Errorf("create share link error: %v", err.Error())
				continue
			}

			sh := sharedLinks[unCompleteTask.OriginalLink]

			if sh == nil {
				log.Errorf("link not found: %s", unCompleteTask.OriginalLink)
				continue
			}

			if sh.State == share.StatusOK {
				now := time.Now()
				link := unCompleteTask.OriginalLink
				s := &model.SharedLink{
					AutoID:             unCompleteTask.AutoID,
					UserID:             unCompleteTask.UserID,
					State:              sh.State.String(),
					Host:               unCompleteTask.Host,
					CreatedBy:          sh.CreatedBy,
					CreatedAt:          unCompleteTask.CreatedAt,
					UpdatedAt:          now,
					Size:               sh.Size,
					Visitor:            sh.Visitor,
					Stored:             sh.Stored,
					Revenue:            sh.Revenue,
					Title:              sh.Title,
					OriginalLinkHash:   lk.Hash(link),
					HostSharedLinkHash: lk.Hash(sh.HostSharedLink),
					OriginalLink:       link,
					HostSharedLink:     sh.HostSharedLink,
				}

				completeTasks = append(completeTasks, s)
				continue
			}

			// if task processing duration grate than 48 hour, it's failed
			if sh.State == share.StatusCreated && time.Now().Sub(sh.CreatedAt).Hours() > 48 {
				failedTasks = append(failedTasks, unCompleteTask)
				continue
			}

			if _, err = query.SharedLink.
				Where(query.SharedLink.AutoID.Eq(unCompleteTask.AutoID)).
				Update(query.SharedLink.UpdatedAt, time.Now()); err != nil {
				log.Errorf("update share link updated_at error: %v", err.Error())
			}
		}

		// update complete tasks state
		if len(completeTasks) > 0 {
			for _, task := range completeTasks {
				if _, err := query.SharedLink.
					Where(query.SharedLink.AutoID.Eq(task.AutoID)).
					Updates(task); err != nil {
					log.WithField("task", task).Errorf("update share link error: %v", err.Error())
				}
			}
		}

		// update failed tasks state
		if len(failedTasks) > 0 {
			for _, task := range failedTasks {
				if _, err := query.SharedLink.
					Where(query.SharedLink.AutoID.Eq(task.AutoID)).
					Update(query.SharedLink.State, share.StatusError); err != nil {
					log.WithField("task", task).Errorf("update share link error: %v", err.Error())
				}
			}
		}
	}
}
