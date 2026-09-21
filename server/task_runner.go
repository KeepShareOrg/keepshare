package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	pm "github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	pq "github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
	"gorm.io/gen/field"
)

type AsyncTaskRunner struct{}

func NewAsyncTaskRunner() *AsyncTaskRunner {
	return &AsyncTaskRunner{}
}

var createNotExistsHostTasksBuffer = make(chan *model.SharedLink, 100000)

func (r *AsyncTaskRunner) ListenCompleteFiles() {
	pp := pq.Use(config.MySQL())
	host := hosts.Get(config.DefaultHost())
	ctx := context.Background()
	host.AddEventListener(hosts.FileComplete, func(userID, originalLinkHash string) {
		log.Debugf("file complete event: %s %s", userID, originalLinkHash)
		files, err := pp.File.WithContext(ctx).Where(
			pp.File.WorkerUserID.Eq(userID),
			pp.File.OriginalLinkHash.Eq(originalLinkHash),
		).Find()
		if err != nil {
			log.Errorf("query files error: %v", err)
			return
		}
		if err := r.handleCompleteUniqueTasks(ctx, files); err != nil {
			log.Errorf("handle complete unique tasks error: %v", err)
		}
	})
}

func (r *AsyncTaskRunner) Run() {
	ctx := context.TODO()
	go r.createNotExistsHostTasks(ctx, createNotExistsHostTasksBuffer)

	for {
		r.WalkDBTasksByState(ctx, []string{
			// just handle pending task, create status have limit conidtion
			share.StatusPending.String(),
		}, func(tasks []*model.SharedLink) error {
			// update by unique_hash
			pp := pq.Use(config.MySQL())
			uniqueHashes := make([]string, 0, len(tasks))
			for _, v := range tasks {
				uniqueHashes = append(uniqueHashes, fmt.Sprintf("%s:%s", v.UserID, v.OriginalLinkHash))
			}
			log.Debugf("uniqueHashes length: %v", len(uniqueHashes))

			ret, err := pp.File.WithContext(ctx).Where(
				pp.File.UniqueHash.In(uniqueHashes...),
				pp.File.Status.In(constant.StatusOK, constant.StatusError),
			).Find()
			if err != nil {
				log.Errorf("query pikpak_file by unique_hash error: %v", err)
				return err
			}

			groupedInfos := lo.GroupBy(ret, func(item *pm.File) string {
				return item.Status
			})
			// handle phase_type_complete
			if v, ok := groupedInfos[constant.StatusOK]; ok && len(v) > 0 {
				log.Debugf("handle complete unique tasks length: %v, %v", len(v), v)
				err := r.handleCompleteUniqueTasks(ctx, lo.Filter(v, func(item *pm.File, _ int) bool {
					return item.Status == constant.StatusOK
				}))
				if err != nil {
					log.Errorf("handle complete unique tasks error: %v", err)
				}
			}
			// handle phase_type_error
			if v, ok := groupedInfos[constant.StatusError]; ok && len(v) > 0 {
				log.Debugf("handle error unique hashes length: %v, %v", len(v), v)
				err := r.handleErrorUniqueHashes(ctx, lo.Map(v, func(item *pm.File, _ int) string {
					return item.UniqueHash
				}))
				if err != nil {
					log.Errorf("handle error unique hashes error: %v", err)
				}
			}

			existsUniqueHashes := lo.Map(ret, func(item *pm.File, _ int) string {
				return item.UniqueHash
			})
			unExistsKeepShareTasks := lo.Filter(tasks, func(item *model.SharedLink, _ int) bool {
				return !lo.Contains(existsUniqueHashes, fmt.Sprintf("%s:%s", item.UserID, item.OriginalLinkHash))
			})
			log.Debugf("unExistsKeepShareTasks: %v", len(unExistsKeepShareTasks))
			// create not exists host tasks
			for _, task := range unExistsKeepShareTasks {
				createNotExistsHostTasksBuffer <- task
			}
			return nil
		})

		time.Sleep(time.Second * 5)
	}
}

// WalkDBTasksByState repeatedly scans tasks in creation order using AutoID to break timestamp ties.
func (r *AsyncTaskRunner) WalkDBTasksByState(ctx context.Context, states []string, fn func(tasks []*model.SharedLink) error) {
	var currentCreatedAt time.Time
	var currentAutoID int64

	for {
		t := query.SharedLink
		stmt := t.WithContext(ctx).Where(t.State.In(states...))
		if currentAutoID > 0 {
			stmt = stmt.Where(
				t.Where(t.CreatedAt.Gt(currentCreatedAt)).
					Or(t.CreatedAt.Eq(currentCreatedAt), t.AutoID.Gt(currentAutoID)),
			)
		}
		ret, err := stmt.Order(t.CreatedAt, t.AutoID).Limit(1000).Find()

		if err != nil {
			log.Errorf("query un complete task err: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(ret) <= 0 {
			currentCreatedAt = time.Time{}
			currentAutoID = 0
			time.Sleep(time.Second)
		} else {
			last := ret[len(ret)-1]
			currentCreatedAt = last.CreatedAt
			currentAutoID = last.AutoID
		}

		if err := fn(ret); err != nil {
			log.Errorf("walk un complete task err: %v", err)
			continue
		}
	}
}

// handleCompleteUniqueTasks handle complete unique tasks
func (r *AsyncTaskRunner) handleCompleteUniqueTasks(ctx context.Context, files []*pm.File) error {
	pp := pq.Use(config.MySQL())
	host := hosts.Get(config.DefaultHost())

	for _, v := range files {
		ok, err := config.Redis().SetNX(ctx, fmt.Sprintf("cs_%s", v.UniqueHash), 1, time.Minute).Result()
		if err != nil {
			log.Errorf("set cs_%s error: %v", v.UniqueHash, err)
			continue
		}
		if !ok {
			log.Debugf("cs_%s already being handled, skip", v.UniqueHash)
			continue
		}

		sharedLink, err := host.CreateShare(ctx, v.MasterUserID, v.WorkerUserID, v.FileID)
		if err != nil {
			log.Errorf("%#v create share error: %v", v, err)
			if IsFileNotFoundError(err) {
				// delete complete pikpak_file
				_, err = pp.File.WithContext(ctx).Where(pp.File.UniqueHash.Eq(v.UniqueHash)).Delete()
				if err != nil {
					log.Errorf("delete pikpak_file error: %v", err)
				}
			}
			continue
		}

		uid, ohs := strings.Split(v.UniqueHash, ":")[0], v.OriginalLinkHash
		if _, err := query.SharedLink.WithContext(ctx).
			Where(
				query.SharedLink.UserID.Eq(uid),
				query.SharedLink.OriginalLinkHash.Eq(ohs),
			).
			Updates(&model.SharedLink{
				State:          share.StatusOK.String(),
				HostSharedLink: sharedLink,
				UpdatedAt:      time.Now(),
			}); err != nil {
			log.Errorf("update keepshare_shared_link state error: %v", err)
			continue
		}
	}

	return nil
}

// handleErrorUniqueHashes update keepshare_shared_link state by unique_hash status
func (r *AsyncTaskRunner) handleErrorUniqueHashes(ctx context.Context, hashes []string) error {
	tupleConditions := lo.Map(hashes, func(hash string, _ int) []string {
		temp := strings.Split(hash, ":")
		if len(temp) != 2 {
			return []string{"", ""}
		}
		uid, ohs := temp[0], temp[1]
		return []string{uid, ohs}
	})

	_, err := query.SharedLink.WithContext(ctx).Where(
		query.SharedLink.WithContext(ctx).
			Columns(query.SharedLink.UserID, query.SharedLink.OriginalLinkHash).
			In(field.Values(tupleConditions)),
	).Updates(&model.SharedLink{
		State:     share.StatusError.String(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Errorf("update keepshare_shared_link state error: %v", err)
		return err
	}

	return nil
}

// createHostTaskIfNotExists create host task if not exists
func (r *AsyncTaskRunner) createNotExistsHostTasks(ctx context.Context, tasks chan *model.SharedLink) {
	eg := errgroup.Group{}
	eg.SetLimit(300)

	for ksl := range tasks {
		eg.Go(func() error {
			rdsKey := fmt.Sprintf("create_not_exists_%s:%s", ksl.UserID, ksl.OriginalLinkHash)
			if ok, _ := config.Redis().SetNX(ctx, rdsKey, 1, time.Minute).Result(); !ok {
				log.Infof("create not exists task is handling by other: %s", ksl.OriginalLink)
				return nil
			}

			defer config.Redis().Del(ctx, rdsKey)

			host := hosts.Get(config.DefaultHost())
			ctx := log.DataContext(ctx, log.DataContextOptions{RequestID: ""})
			log := log.WithContext(ctx)
			log.Infof("should create file: %#v %v", ksl, host)
			sharedLinks, err := host.CreateFromLinks(ctx, ksl.UserID, []string{ksl.OriginalLink}, ksl.CreatedBy, "")
			if err != nil {
				log.WithFields(map[string]interface{}{
					"user_id":       ksl.UserID,
					"original_link": ksl.OriginalLink,
				}).Debugf("create share from links err: %v", err)

				log.Errorf("create share from links err: %v", err)

				if _, err = query.SharedLink.WithContext(ctx).
					Where(query.SharedLink.AutoID.Eq(ksl.AutoID)).
					Updates(&model.SharedLink{
						State:     share.StatusError.String(),
						Error:     err.Error(),
						UpdatedAt: time.Now(),
					}); err != nil {
					log.Errorf("update keepshare_shared_link state error: %v", err)
				}
			} else {
				log.Debugf("create share from links ok: %s", ksl.OriginalLink)
				// update keepshare shared link info
				sh, ok := sharedLinks[ksl.OriginalLink]
				if ok {
					if _, err = query.SharedLink.WithContext(ctx).
						Where(query.SharedLink.AutoID.Eq(ksl.AutoID)).
						Updates(&model.SharedLink{
							State:     sh.State.String(),
							Title:     sh.Title,
							Size:      sh.Size,
							UpdatedAt: time.Now(),
						}); err != nil {
						log.Errorf("update keepshare_shared_link state error: %v", err)
					}
				}
			}

			return nil
		})
	}

	_ = eg.Wait()
}
