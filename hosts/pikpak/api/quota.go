// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	log "github.com/sirupsen/logrus"
)

// GetStorageSize returns the used and limit size of the worker.
func (api *API) GetStorageSize(ctx context.Context, worker string) (used, limit int64, err error) {
	token, err := api.getToken(ctx, worker)
	if err != nil {
		return 0, 0, err
	}

	var e RespErr
	var r struct {
		Quota struct {
			Limit string `json:"limit"`
			Usage string `json:"usage"`
		} `json:"quota"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		Get(apiURL("/drive/v1/about"))

	if err != nil {
		return 0, 0, fmt.Errorf("get storage err: %w", err)
	}

	log.WithField("worker", worker).Debugf("get storage resp body: %s", body.Body())

	if err = e.Error(); err != nil {
		return 0, 0, fmt.Errorf("get storage err: %w", err)
	}

	if r.Quota.Limit == "" {
		return 0, 0, fmt.Errorf("get storage got unexpected body: %s", body.Body())
	}

	used, _ = strconv.ParseInt(r.Quota.Usage, 10, 64)
	limit, _ = strconv.ParseInt(r.Quota.Limit, 10, 64)
	return used, limit, nil
}

var defaultExpiration, _ = time.ParseInLocation("2006-01-02", "2000-01-01", time.Local)

// GetPremiumExpiration get the premium expiration for the worker account.
func (api *API) GetPremiumExpiration(ctx context.Context, worker string) (*time.Time, error) {
	token, err := api.getToken(ctx, worker)
	if err != nil {
		return nil, err
	}

	var e RespErr
	var r struct {
		Data struct {
			Expire string `json:"expire"`
			Status string `json:"status"`
			Type   string `json:"type"`
		} `json:"data"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		Get(apiURL("/drive/v1/privilege/vip"))

	if err != nil {
		return nil, fmt.Errorf("get premium expiration err: %w", err)
	}

	log.WithField("worker", worker).Debugf("get premium expiration resp body: %s", body.Body())

	if err = e.Error(); err != nil {
		return nil, fmt.Errorf("get premium expiration err: %w", err)
	}

	if r.Data.Status == "" {
		return nil, fmt.Errorf("get premium expiration got unexpected body: %s", body.Body())
	}

	t, err := time.Parse(time.RFC3339, r.Data.Expire)
	if err != nil || t.Before(defaultExpiration) {
		return &defaultExpiration, nil
	}

	return &t, nil
}

func (api *API) updatePremiumExpirationBackground() {
	// TODO
}

// UpdateWorkerStorage updates the worker's storage info from server.
func (api *API) UpdateWorkerStorage(ctx context.Context, worker string) error {
	t := &api.q.WorkerAccount
	used, limit, err := api.GetStorageSize(ctx, worker)
	if err != nil {
		return fmt.Errorf("get storage size err: %w", err)
	}

	w := &model.WorkerAccount{
		UserID:    worker,
		UsedSize:  used,
		LimitSize: limit,
		UpdatedAt: time.Now(),
	}
	_, err = t.WithContext(ctx).Updates(w)
	log.Debugf("worker info updated: %+v", w)
	return err
}

// UpdateWorkerPremium updates the worker's premium expiration and also storage info.
func (api *API) UpdateWorkerPremium(ctx context.Context, worker *model.WorkerAccount) error {
	// update the worker's premium expiration once in 24h.
	key := fmt.Sprintf("pikpak:updateWorkerPremium:%s", worker.UserID)
	ok, _ := api.Redis.SetNX(ctx, key, "", 24*time.Hour).Result()
	if !ok {
		log.Debugf("ignore update worker")
		return nil
	}

	t := &api.q.WorkerAccount
	exp, err := api.GetPremiumExpiration(ctx, worker.UserID)
	if err != nil {
		return fmt.Errorf("get premium expiration err: %w", err)
	}

	used, limit, err := api.GetStorageSize(ctx, worker.UserID)
	if err != nil {
		return fmt.Errorf("get storage size err: %w", err)
	}

	if worker.PremiumExpiration.Unix() == exp.Unix() && worker.UsedSize == used && worker.LimitSize == limit {
		// nothing changed
		log.Debugf("worker info not changed")
		return nil
	}

	worker.PremiumExpiration = *exp
	worker.UsedSize = used
	worker.LimitSize = limit
	worker.UpdatedAt = time.Now()

	_, err = t.WithContext(ctx).Select(t.PremiumExpiration, t.UsedSize, t.LimitSize, t.UpdatedAt).Updates(worker)
	log.Debugf("worker info updated: %+v", worker)
	return err
}
