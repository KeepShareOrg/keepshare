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
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	log "github.com/sirupsen/logrus"
)

// CreateShare create a sharing link by files.
func (api *API) CreateShare(ctx context.Context, master string, worker string, fileID string) (sharedLink string, err error) {
	token, err := api.getToken(ctx, worker)
	if err != nil {
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

	if err = e.Error(); err != nil {
		// TODO token expired
		return "", fmt.Errorf("create share err: %w", err)
	}

	if r.SharedLink == "" {
		return "", fmt.Errorf("unexpected body: %s", body.Body())
	}

	id, _ := getShareIDFromLink(r.SharedLink)

	err = api.q.SharedLink.WithContext(ctx).Create(&model.SharedLink{
		ShareID:      id,
		FileID:       fileID,
		MasterUserID: master,
		WorkerUserID: worker,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return "", err
	}
	return r.SharedLink, nil
}

// GetShareStatus returns the statuses of the shared link.
func (api *API) GetShareStatus(ctx context.Context, sharedLink string) (status share.State, worker string, err error) {
	id, err := getShareIDFromLink(sharedLink)
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

	log.WithFields(map[string]any{
		constant.SharedLink:  sharedLink,
		constant.ShareStatus: status,
	}).Debugf("get status `%s` from server", r.ShareStatus)

	return status, r.UserInfo.UserID, nil
}

// GetStatistics returns the statistics of each shared links.
func (api *API) GetStatistics(ctx context.Context, sharedLink string) (*share.Statistics, error) {
	id, err := getShareIDFromLink(sharedLink)
	if err != nil {
		return nil, err
	}

	t := &api.q.SharedLink
	sh, err := t.WithContext(ctx).Where(t.ShareID.Eq(id)).Take()
	if err != nil {
		return nil, err
	}

	token, err := api.getToken(ctx, sh.WorkerUserID)
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

	log.WithFields(map[string]any{
		constant.SharedLink:  sharedLink,
		constant.ShareStatus: r.Data[0].ShareStatus,
		"statistics":         st,
	}).Debug("")

	return st, nil
}

func getShareIDFromLink(link string) (id string, err error) {
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

// DeleteShare create a sharing link by files.
func (api *API) DeleteShare(ctx context.Context, worker string, shareIDs []string) error {
	if len(shareIDs) == 0 {
		return nil
	}

	token, err := api.getToken(ctx, worker)
	if err != nil {
		return err
	}

	var e RespErr
	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetBody(JSON{"ids": shareIDs}).
		Post(apiURL("/drive/v1/share:batchDelete"))

	if err != nil {
		return fmt.Errorf("delete share from server err: %w", err)
	}

	if err = e.Error(); err != nil {
		// TODO token expired
		return fmt.Errorf("delete share from server err: %w", err)
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		log.WithFields(map[string]any{
			"worker":    worker,
			"share_ids": shareIDs,
		}).Debugf("delete share resp body: %s", body)
	}
	_, err = api.q.SharedLink.WithContext(ctx).Where(api.q.SharedLink.ShareID.In(shareIDs...)).Delete()
	if err != nil {
		return fmt.Errorf("delete share from db err: %w", err)
	}
	return nil
}

// DeleteShareByFileIDs deletes host shared links by original links.
func (api *API) DeleteShareByFileIDs(ctx context.Context, worker string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return nil
	}
	t := &api.q.SharedLink
	rows, err := t.WithContext(ctx).Select(t.ShareID).Where(t.FileID.In(fileIDs...)).Find()
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
