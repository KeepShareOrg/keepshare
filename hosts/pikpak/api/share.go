// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/magnetblock"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

// CreateShare create a sharing link by files.
func (api *API) CreateShare(ctx context.Context, master string, worker string, fileID string) (sharedLink string, err error) {
	var hashes []string
	if err := api.Mysql.WithContext(ctx).Clauses(dbresolver.Write).Model(&model.File{}).
		Where("worker_user_id = ? AND file_id = ?", worker, fileID).
		Pluck("original_link_hash", &hashes).Error; err != nil {
		return "", fmt.Errorf("get share source hashes: %w", err)
	}
	if len(hashes) == 0 {
		return "", errors.New("share source file not found")
	}
	if err := api.checkShareHashes(ctx, hashes); err != nil {
		return "", err
	}

	token, err := api.getToken(ctx, worker, false)
	if err != nil {
		return "", err
	}
	if err := api.checkShareHashes(ctx, hashes); err != nil {
		return "", err
	}

	var e RespErr
	var r struct {
		SharedLink string `json:"share_url"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetBody(JSON{
			"file_ids":         []string{fileID},
			"share_to":         "publiclink",
			"expiration_days":  -1,
			"pass_code_option": "NOT_REQUIRED",
		}).
		Post(apiURL("/drive/v1/share"))

	if err != nil {
		return "", fmt.Errorf("create share err: %w", err)
	}
	if !body.IsSuccess() {
		return "", fmt.Errorf("create share failed: HTTP %d", body.StatusCode())
	}

	if err = e.Error(); err != nil {
		// TODO token expired
		return "", fmt.Errorf("create share err: %w", err)
	}

	if r.SharedLink == "" {
		return "", fmt.Errorf("unexpected body: %s", body.Body())
	}

	id, err := ShareIDFromLink(r.SharedLink)
	if err != nil {
		return "", fmt.Errorf("create share returned an invalid URL: %w", err)
	}

	// The remote share must remain discoverable even when the caller is canceled.
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	err = api.q.SharedLink.WithContext(cleanupCtx).Clauses(clause.OnConflict{DoNothing: true}).Create(&model.SharedLink{
		ShareID:      id,
		FileID:       fileID,
		MasterUserID: master,
		WorkerUserID: worker,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return "", api.discardCreatedShare(ctx, worker, id, fmt.Errorf("store created share mapping: %w", err))
	}
	if err := api.checkShareHashes(cleanupCtx, hashes); err != nil {
		return "", api.discardCreatedShare(ctx, worker, id, err)
	}
	return r.SharedLink, nil
}

// GetShareStatus returns the statuses of the shared link.
func (api *API) GetShareStatus(ctx context.Context, sharedLink string) (status share.State, worker string, err error) {
	id, err := ShareIDFromLink(sharedLink)
	if err != nil {
		return share.StatusUnknown, "", err
	}

	var e RespErr
	var r struct {
		ShareStatus string `json:"share_status"`
		UserInfo    struct {
			UserID string `json:"user_id"`
		} `json:"user_info"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetError(&e).
		SetResult(&r).
		SetQueryParams(map[string]string{"share_id": id}).
		Get(apiURL("/drive/v1/share"))

	if err != nil {
		return "", "", fmt.Errorf("get share status err: %w", err)
	}

	if err = e.Error(); err != nil {
		return "", "", fmt.Errorf("get share status err: %w", err)
	}

	if r.ShareStatus == "" || r.UserInfo.UserID == "" {
		return "", "", fmt.Errorf("unexpected body: %s", body.Body())
	}

	switch r.ShareStatus {
	case "OK":
		status = share.StatusOK
	case "DELETED":
		status = share.StatusDeleted
	case "NOT_FOUND":
		status = share.StatusNotFound
	case "SENSITIVE_RESOURCE", "SENSITIVE_WORD":
		status = share.StatusSensitive
	default:
		status = share.StatusUnknown
	}

	log.WithContext(ctx).WithFields(map[string]any{
		constant.SharedLink:  sharedLink,
		constant.ShareStatus: status,
	}).Debugf("get status `%s` from server", r.ShareStatus)

	return status, r.UserInfo.UserID, nil
}

// GetStatistics returns the statistics of each shared links.
func (api *API) GetStatistics(ctx context.Context, sharedLink string) (*share.Statistics, error) {
	id, err := ShareIDFromLink(sharedLink)
	if err != nil {
		return nil, err
	}

	t := &api.q.SharedLink
	sh, err := t.WithContext(ctx).Where(t.ShareID.Eq(id)).Take()
	if err != nil {
		return nil, err
	}

	token, err := api.getToken(ctx, sh.WorkerUserID, false)
	if err != nil {
		return nil, err
	}

	var e RespErr
	var r struct {
		Data []struct {
			RestoreCount string `json:"restore_count"`
			ShareID      string `json:"share_id"`
			ShareStatus  string `json:"share_status"`
			ViewCount    string `json:"view_count"`
		} `json:"data"`
	}

	filters := map[string]any{
		"id": map[string]any{
			"in": id,
		},
	}
	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetQueryParams(map[string]string{"filters": util.ToJSON(filters)}).
		Get(apiURL("/drive/v1/share/list"))

	if err != nil {
		return nil, fmt.Errorf("get share statistics err: %w", err)
	}

	if err = e.Error(); err != nil {
		return nil, fmt.Errorf("get share statistics err: %w", err)
	}

	if len(r.Data) == 0 || r.Data[0].ShareID != id {
		return nil, fmt.Errorf("unexpected body: %s", body.Body())
	}

	//if r.Data[0].ShareStatus != "OK" {
	//	// TODO update shared link?
	//}

	st := &share.Statistics{
		Visitor: int32(util.Atoi(r.Data[0].ViewCount)),
		Stored:  int32(util.Atoi(r.Data[0].RestoreCount)),
		Revenue: 0, // TODO
	}

	log.WithContext(ctx).WithFields(map[string]any{
		constant.SharedLink:  sharedLink,
		constant.ShareStatus: r.Data[0].ShareStatus,
		"statistics":         st,
	}).Debug("")

	return st, nil
}

// ShareIDFromLink extracts the native share identity, excluding URL action parameters.
func ShareIDFromLink(link string) (id string, err error) {
	u, _ := url.Parse(link)
	if u == nil || !strings.HasPrefix(u.Path, "/s/") {
		return "", errors.New("invalid link")
	}

	id, _, _ = strings.Cut(strings.TrimPrefix(u.Path, "/s/"), "/")
	if id == "" {
		return "", errors.New("invalid link")
	}

	return id, nil
}

// DeleteShare revokes native shares and removes only their local share mappings.
func (api *API) DeleteShare(ctx context.Context, worker string, shareIDs []string) error {
	if len(shareIDs) == 0 {
		return nil
	}
	if err := api.deleteShareRemote(ctx, worker, shareIDs); err != nil {
		return err
	}
	_, err := api.q.SharedLink.WithContext(ctx).Where(
		api.q.SharedLink.WorkerUserID.Eq(worker),
		api.q.SharedLink.ShareID.In(shareIDs...),
	).Delete()
	if err != nil {
		return fmt.Errorf("delete share from db err: %w", err)
	}
	return nil
}

func (api *API) deleteShareRemote(ctx context.Context, worker string, shareIDs []string) error {
	if len(shareIDs) == 0 {
		return nil
	}

	token, err := api.getToken(ctx, worker, false)
	if err != nil {
		// Credential responses can contain tokens; do not include them in CLI errors.
		return errors.New("worker credential lookup or refresh failed")
	}

	var e RespErr
	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&e).
		SetBody(JSON{"ids": shareIDs}).
		Post(apiURL("/drive/v1/share:batchDelete"))

	if err != nil {
		return fmt.Errorf("delete share from server err: %w", err)
	}
	if !body.IsSuccess() {
		return fmt.Errorf("delete share failed: HTTP %d (code %d)", body.StatusCode(), e.ErrorCode)
	}

	if e.ErrorCode > 0 || e.Error() != nil {
		return fmt.Errorf("delete share rejected (code %d)", e.ErrorCode)
	}

	if log.IsDebugEnabled() {
		log.WithContext(ctx).WithFields(map[string]any{
			"worker":    worker,
			"share_ids": shareIDs,
		}).Debugf("delete share resp body: %s", body)
	}
	return nil
}

// DeleteShareByFileIDs deletes host shared links by original links.
func (api *API) DeleteShareByFileIDs(ctx context.Context, worker string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return nil
	}
	t := &api.q.SharedLink
	rows, err := t.WithContext(ctx).Select(t.ShareID).Where(
		t.WorkerUserID.Eq(worker),
		t.FileID.In(fileIDs...),
	).Find()
	if err != nil && !gormutil.IsNotFoundError(err) {
		return fmt.Errorf("select share ids from db err: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	shareIDs := make([]string, 0, len(rows))
	for _, v := range rows {
		shareIDs = append(shareIDs, v.ShareID)
	}
	return api.DeleteShare(ctx, worker, shareIDs)
}

func (api *API) checkShareHashes(ctx context.Context, hashes []string) error {
	blocked, err := magnetblock.FindBlocked(ctx, api.Mysql, hashes)
	if err != nil {
		return err
	}
	for _, hash := range hashes {
		if blocked[hash] {
			return magnetblock.ErrBlocked
		}
	}
	return nil
}

func (api *API) discardCreatedShare(ctx context.Context, worker, id string, cause error) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	return errors.Join(cause, api.DeleteShare(cleanupCtx, worker, []string{id}))
}
