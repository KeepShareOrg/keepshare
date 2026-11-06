// Copyright 2026 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/magnetblock"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	keepshare "github.com/KeepShareOrg/keepshare/server/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

var shareIDLikeEscaper = strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")

// RevokeResult counts confirmed remote revocations, including local cleanup failures.
type RevokeResult struct {
	Revoked  int             `json:"revoked"`
	Failures []RevokeFailure `json:"failures"`
}

// RevokeFailure identifies a failed share operation without credential contents.
type RevokeFailure struct {
	MasterUserID string `json:"master_user_id,omitempty"`
	WorkerUserID string `json:"worker_user_id,omitempty"`
	ShareID      string `json:"share_id,omitempty"`
	Stage        string `json:"stage"`
	Error        string `json:"error"`
}

// RevokeMagnetShares revokes native shares across all accounts, never cloud files.
func (api *API) RevokeMagnetShares(ctx context.Context, link string) (RevokeResult, error) {
	result := RevokeResult{Failures: make([]RevokeFailure, 0)}
	hash, err := magnetblock.Normalize(link)
	if err != nil {
		return result, err
	}
	aliases := magnetblock.Aliases(hash)
	db := api.Mysql.WithContext(ctx).Clauses(dbresolver.Write).Session(&gorm.Session{})

	var fileShares []*model.SharedLink
	if err := db.Table(model.TableNameSharedLink+" AS s").Select("DISTINCT s.*").
		Joins("JOIN "+model.TableNameFile+" AS f ON f.worker_user_id = s.worker_user_id AND f.file_id = s.file_id").
		Where("f.original_link_hash IN ?", aliases).Find(&fileShares).Error; err != nil {
		return result, fmt.Errorf("find native shares by magnet files: %w", err)
	}

	// File records may already be gone while a Keepshare URL and native mapping remain.
	var records []*keepshare.SharedLink
	if err := db.Where("host = ? AND original_link_hash IN ? AND host_shared_link <> ''", "pikpak", aliases).
		Find(&records).Error; err != nil {
		return result, fmt.Errorf("find historical Keepshare URLs: %w", err)
	}
	urlIDs := make(map[string]bool)
	for _, record := range records {
		if id, err := ShareIDFromLink(record.HostSharedLink); err == nil {
			urlIDs[id] = true
		}
	}
	ids := make([]string, 0, len(urlIDs))
	for id := range urlIDs {
		ids = append(ids, id)
	}
	var urlShares []*model.SharedLink
	if len(ids) > 0 {
		if err := db.Where("share_id IN ?", ids).Find(&urlShares).Error; err != nil {
			return result, fmt.Errorf("find historical native share mappings: %w", err)
		}
	}

	targets := make(map[string]*model.SharedLink, len(fileShares)+len(urlShares))
	for _, item := range fileShares {
		targets[item.ShareID] = item
	}
	for _, item := range urlShares {
		targets[item.ShareID] = item
	}
	sort.Strings(ids)
	for _, id := range ids {
		if targets[id] == nil {
			result.Failures = append(result.Failures, RevokeFailure{
				ShareID: id, Stage: "mapping", Error: "native share account mapping not found",
			})
		}
	}

	byWorker := make(map[string][]*model.SharedLink)
	for _, item := range targets {
		byWorker[item.WorkerUserID] = append(byWorker[item.WorkerUserID], item)
	}
	workers := make([]string, 0, len(byWorker))
	for worker := range byWorker {
		workers = append(workers, worker)
	}
	sort.Strings(workers)
	for _, worker := range workers {
		items := byWorker[worker]
		sort.Slice(items, func(i, j int) bool { return items[i].ShareID < items[j].ShareID })
		shareIDs := make([]string, 0, len(items))
		for _, item := range items {
			shareIDs = append(shareIDs, item.ShareID)
		}
		if err := api.deleteShareRemote(ctx, worker, shareIDs); err != nil {
			for _, item := range items {
				result.addFailure(item, "revoke", err)
			}
			continue
		}
		result.Revoked += len(items)
		for _, item := range items {
			if stage, err := api.clearRevokedShare(ctx, item); err != nil {
				result.addFailure(item, stage, err)
			}
		}
	}
	if len(result.Failures) > 0 {
		return result, fmt.Errorf("magnet share revocation has %d failed operation(s)", len(result.Failures))
	}
	return result, nil
}

func (r *RevokeResult) addFailure(item *model.SharedLink, stage string, err error) {
	r.Failures = append(r.Failures, RevokeFailure{
		MasterUserID: item.MasterUserID,
		WorkerUserID: item.WorkerUserID,
		ShareID:      item.ShareID,
		Stage:        stage,
		Error:        err.Error(),
	})
}

// Lock the native mapping before finding consumers so delayed publishers cannot escape cleanup.
func (api *API) clearRevokedShare(ctx context.Context, item *model.SharedLink) (string, error) {
	db := api.Mysql.WithContext(ctx).Clauses(dbresolver.Write).Session(&gorm.Session{})
	stage := "cleanup"
	err := db.Transaction(func(tx *gorm.DB) error {
		tx = tx.Session(&gorm.Session{})
		var native model.SharedLink
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("worker_user_id = ? AND share_id = ?", item.WorkerUserID, item.ShareID).Take(&native).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		var candidates []*keepshare.SharedLink
		escapedID := shareIDLikeEscaper.Replace(item.ShareID)
		if err := tx.Where("host = ? AND host_shared_link LIKE ?", "pikpak", "%/s/"+escapedID+"%").
			Find(&candidates).Error; err != nil {
			return fmt.Errorf("find revoked Keepshare URLs: %w", err)
		}
		records := candidates[:0]
		cacheKeys := make([]string, 0, len(candidates))
		for _, record := range candidates {
			if id, err := ShareIDFromLink(record.HostSharedLink); err == nil && id == item.ShareID {
				records = append(records, record)
				cacheKeys = append(cacheKeys, fmt.Sprintf("status:%d", record.AutoID))
			}
		}
		if len(cacheKeys) > 0 {
			stage = "cache"
			if api.Redis == nil {
				return errors.New("Redis is required to clear revoked share status caches")
			}
			if err := api.Redis.Del(ctx, cacheKeys...).Err(); err != nil {
				return fmt.Errorf("clear revoked share status cache: %w", err)
			}
		}
		stage = "cleanup"
		now := time.Now()
		for _, record := range records {
			if err := tx.Model(&keepshare.SharedLink{}).
				Where("auto_id = ? AND host_shared_link = ?", record.AutoID, record.HostSharedLink).
				Updates(map[string]any{
					"host_shared_link":      "",
					"host_shared_link_hash": "",
					"state": gorm.Expr("CASE WHEN state = ? THEN state ELSE ? END",
						share.StatusBlocked.String(), share.StatusDeleted.String()),
					"updated_at": now,
				}).Error; err != nil {
				return err
			}
		}
		return tx.Where("worker_user_id = ? AND share_id = ?", item.WorkerUserID, item.ShareID).
			Delete(&model.SharedLink{}).Error
	})
	if err != nil {
		return stage, fmt.Errorf("clear revoked share records: %w", err)
	}
	return "", nil
}
