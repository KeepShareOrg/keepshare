package server

import (
	"context"
	"fmt"
	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	pq "github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"gorm.io/gorm/utils"
	"time"
)

type AsyncTaskRunner struct{}

func NewAsyncTaskRunner() *AsyncTaskRunner {
	return &AsyncTaskRunner{}
}

func (r *AsyncTaskRunner) Run() {
	ctx := context.TODO()
	r.WalkDBTasksByState(ctx, []string{
		share.StatusCreated.String(),
	}, func(tasks []*model.SharedLink) error {
		// update by unique_hash
		uniqueHashes := make([]string, 0, len(tasks))
		for _, v := range tasks {
			uniqueHashes = append(uniqueHashes, fmt.Sprintf("%s:%s", v.UserID, v.OriginalLinkHash))
		}
		log.Debugf("uniqueHashes length: %v", len(uniqueHashes))

		if err := r.updateStateByUniqueHashStatus(ctx, uniqueHashes); err != nil {
			log.Errorf("update state by original link hash status error: %v", err)
			return nil
		}

		if err := r.createHostTaskIfNotExists(ctx, uniqueHashes); err != nil {
			log.Errorf("create host task if not exists error: %v", err)
			return err
		}
		return nil
	})
}

// WalkDBTasksByState walk db tasks
func (r *AsyncTaskRunner) WalkDBTasksByState(ctx context.Context, states []string, fn func(tasks []*model.SharedLink) error) {
	var currentAutoID int64 = 0

	for {
		ret, err := query.SharedLink.WithContext(gormutil.IgnoreTraceContext(ctx)).
			Where(
				query.SharedLink.AutoID.Gt(currentAutoID),
				query.SharedLink.State.In(states...),
			).Order(query.SharedLink.AutoID).
			Limit(5000).
			Find()

		if err != nil {
			log.Errorf("query un complete task err: %v", err)
			return
		}

		if len(ret) <= 0 {
			currentAutoID = 0
			time.Sleep(time.Second)
		} else {
			currentAutoID = ret[len(ret)-1].AutoID
		}

		if err := fn(ret); err != nil {
			log.Errorf("walk un complete task err: %v", err)
			break
		}

	}
}

// updateStateByUniqueHashStatus update keepshare_shared_link state by unique_hash status
func (r *AsyncTaskRunner) updateStateByUniqueHashStatus(ctx context.Context, uniqueHashes []string) error {
	pp := pq.Use(config.MySQL())

	ret, err := pp.File.WithContext(ctx).Where(
		pp.File.UniqueHash.In(uniqueHashes...),
		pp.File.Status.In(constant.StatusOK, constant.StatusError),
	).Find()
	if err != nil {
		log.Errorf("query pikpak_file by unique_hash error: %v", err)
		return err
	}

	for _, v := range ret {
		if !utils.Contains([]string{constant.StatusOK, constant.StatusError}, v.Status) {
			continue
		}

		info, err := pp.MasterAccount.WithContext(ctx).
			Where(pp.MasterAccount.UserID.Eq(v.MasterUserID)).
			Take()
		if err != nil {
			log.Errorf("query master account info error: %v", err)
			continue
		}

		targetLink := query.SharedLink.WithContext(ctx).
			Where(
				query.SharedLink.UserID.Eq(info.KeepshareUserID),
				query.SharedLink.OriginalLinkHash.Eq(v.OriginalLinkHash),
			)

		// update ok original_link_hash state
		if v.Status == constant.StatusOK {
			// create share link
			host := hosts.Get(config.DefaultHost())
			sharedLink, err := host.CreateShare(ctx, v.MasterUserID, v.WorkerUserID, v.FileID)
			if err != nil {
				log.Errorf("%#v create share error: %v", v, err)
				continue
			}
			if _, err := targetLink.
				Updates(&model.SharedLink{
					State:          share.StatusOK.String(),
					HostSharedLink: sharedLink,
					UpdatedAt:      time.Now(),
				}); err != nil {
				log.Errorf("update keepshare_shared_link state error: %v", err)
				continue
			}
		}

		// update error original_link_hash state
		if v.Status == constant.StatusError {
			if _, err := targetLink.Update(query.SharedLink.State, share.StatusError.String()); err != nil {
				log.Errorf("update keepshare_shared_link state error: %v", err)
				continue
			}
		}
	}

	return nil
}

// createHostTaskIfNotExists create host task if not exists
func (r *AsyncTaskRunner) createHostTaskIfNotExists(ctx context.Context, originalLinkHashes []string) error {
	// TODO: query pikpak_file, if task not in pikpak_file create host task
	return nil
}
