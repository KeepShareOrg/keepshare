// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"gorm.io/gorm/clause"
)

func (api *API) getToken(ctx context.Context, userID string) (token string, err error) {
	key := []byte("token:" + userID)
	ttl := 2 * time.Minute

	v, _ := api.cache.Get(key)
	if len(v) > 0 {
		return string(v), nil
	}

	defer func() {
		if err == nil && token != "" {
			api.cache.Set(key, []byte(token), int(ttl/time.Second))
		}
	}()

	r, err := api.q.Token.WithContext(ctx).Where(api.q.Token.UserID.Eq(userID)).Take()
	if err != nil && !gormutil.IsNotFoundError(err) {
		return "", fmt.Errorf("query token err: %w", err)
	}

	if r != nil && r.Expiration.After(time.Now().Add(ttl+20*time.Second)) {
		return r.AccessToken, nil
	}

	return api.createToken(ctx, userID)
}

func (api *API) createToken(ctx context.Context, userID string) (string, error) {
	account, err := api.q.WorkerAccount.WithContext(ctx).Where(api.q.WorkerAccount.UserID.Eq(userID)).Take()
	if err != nil {
		return "", fmt.Errorf("get worker account err: %w", err)
	}

	var e RespErr
	var r struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"` // TODO
		ExpiresIn    int64  `json:"expires_in"`
	}
	body, err := resCli.R().
		SetContext(ctx).
		SetResult(&r).
		SetError(&e).
		SetBody(JSON{
			"username":  account.Email,
			"password":  account.Password,
			"client_id": clientID,
		}).
		Post(userURL("/v1/auth/signin"))

	if err != nil {
		return "", fmt.Errorf("sign in err: %w", err)
	}

	if err := e.Error(); err != nil {
		return "", fmt.Errorf("sign in err: %w", err)
	}

	if r.AccessToken == "" {
		return "", fmt.Errorf("unexpected body: %s", body.Body())
	}

	now := time.Now()
	t := &model.Token{
		UserID:       userID,
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		Expiration:   now.Add(time.Duration(r.ExpiresIn) * time.Second),
		CreatedAt:    now,
	}

	err = api.q.Token.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(t)
	if err != nil {
		return "", fmt.Errorf("update token err: %w", err)
	}

	return t.AccessToken, nil
}
