// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"gorm.io/gorm/clause"
)

func (api *API) getToken(ctx context.Context, userID string, isMaster bool) (token string, err error) {
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

	token, err = api.createToken(ctx, userID, isMaster)
	if IsInvalidAccountOrPasswordErr(err) {
		log.WithContext(ctx).Debugf("delete invalid account: %s", userID)
		api.q.WorkerAccount.WithContext(ctx).Where(api.q.WorkerAccount.UserID.Eq(userID)).Delete()
	}

	return token, err
}

func (api *API) createToken(ctx context.Context, userID string, isMaster bool) (string, error) {
	var email, password string
	if isMaster {
		account, err := api.q.MasterAccount.WithContext(ctx).Where(api.q.MasterAccount.UserID.Eq(userID)).Take()
		if err != nil {
			return "", fmt.Errorf("get master account err: %w", err)
		}
		email, password = account.Email, account.Password
	} else {
		account, err := api.q.WorkerAccount.WithContext(ctx).Where(api.q.WorkerAccount.UserID.Eq(userID)).Take()
		if err != nil {
			return "", fmt.Errorf("get worker account err: %w", err)
		}
		email, password = account.Email, account.Password
	}

	log.WithContext(ctx).Debugf("create token for email: %s password: %s", email, password)

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
			"username":  email,
			"password":  password,
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
