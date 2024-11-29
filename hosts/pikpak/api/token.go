// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/util"
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

	if r != nil && r.RefreshToken != "" {
		tokenInfo, err := api.RefreshToken(ctx, userID, r.RefreshToken)
		if err == nil {
			return tokenInfo.AccessToken, nil
		}
		api.q.Token.WithContext(ctx).Where(api.q.Token.UserID.Eq(userID)).Delete()
	}

	token, err = api.CreateToken(ctx, userID, isMaster)
	if IsEmptyPasswordErr(err) {
		if isMaster {
			// TODO: should notify the user, fill the new password or reset password
		}
	} else if IsInvalidAccountOrPasswordErr(err) {
		log.WithContext(ctx).Debugf("delete invalid account: %s", userID)
		if isMaster {
			ma := api.q.MasterAccount
			ma.WithContext(ctx).Where(ma.UserID.Eq(userID)).Update(ma.Password, "")
		} else {
			wa := api.q.WorkerAccount
			wa.WithContext(ctx).Where(wa.UserID.Eq(userID)).Delete()
		}
	}

	return token, err
}

func (api *API) CreateToken(ctx context.Context, userID string, isMaster bool) (string, error) {
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

	// if the user does not check the remember me checkbox, password is empty
	if password == "" {
		return "", EmptyPasswordErr
	}

	tokenInfo, err := api.signIn(ctx, userID, email, password)
	if err != nil {
		return "", err
	}

	return tokenInfo.AccessToken, nil
}

func (api *API) RefreshToken(ctx context.Context, userID, refreshToken string) (*model.Token, error) {
	var e *RespErr
	var r struct {
		TokenType    string   `json:"token_type"`
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		IdToken      string   `json:"id_token"`
		ExpiresIn    int      `json:"expires_in"`
		Scope        string   `json:"scope"`
		Sub          string   `json:"sub"`
		UserGroup    []string `json:"user_group"`
		UserId       string   `json:"user_id"`
	}

	resp, err := resCli.R().SetContext(ctx).SetError(&e).SetResult(&r).SetBody(JSON{
		"client_id":     clientID,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}).Post(userURL("/v1/auth/token"))

	log.Debugf("refresh token resp body: %v", resp)

	if err := e.Error(); err != nil {
		return nil, fmt.Errorf("refresh token err: %w", err)
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
		return nil, fmt.Errorf("update token err: %w", err)
	}

	return t, nil
}

func (api *API) ConfirmPassword(ctx context.Context, keepShareUserID, password string, savePassword bool) error {
	ma := api.q.MasterAccount
	account, err := ma.WithContext(ctx).Where(ma.KeepshareUserID.Eq(keepShareUserID)).Take()
	if err != nil {
		return fmt.Errorf("get master account info err: %w", err)
	}

	_, err = api.signIn(ctx, account.UserID, account.Email, password)
	if err != nil {
		return fmt.Errorf("sign in err: %w", err)
	}

	if savePassword {
		ma.WithContext(ctx).Where(ma.KeepshareUserID.Eq(keepShareUserID)).Update(ma.Password, password)
	}
	return nil
}

type signInResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (api *API) signIn(ctx context.Context, userID, username, password string) (*signInResponse, error) {
	var e *RespErr
	var r *signInResponse

	captchaToken, err := api.signInCaptcha(ctx, username)
	if err != nil {
		return nil, err
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetResult(&r).
		SetError(&e).
		SetHeader("X-Captcha-Token", captchaToken).
		SetBody(JSON{
			"username":  username,
			"password":  password,
			"client_id": clientID,
		}).
		Post(userURL("/v1/auth/signin"))

	if err != nil {
		return nil, fmt.Errorf("sign in err: %w", err)
	}

	if err := e.Error(); err != nil {
		return nil, fmt.Errorf("sign in err: %w", err)
	}

	if r.AccessToken == "" {
		return nil, fmt.Errorf("unexpected body: %s", body.Body())
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
		return nil, fmt.Errorf("update token err: %w", err)
	}

	return r, nil
}

func (api *API) signInCaptcha(ctx context.Context, email string) (string, error) {
	var r struct {
		*RespErr
		CaptchaToken string `json:"captcha_token"`
	}

	b := util.ToJSON(map[string]any{
		"action":    "POST:/v1/auth/signin",
		"client_id": webClientID,
		"device_id": deviceID,
		"meta": map[string]string{
			"email": email,
		},
	})
	resp, err := resCli.R().SetContext(ctx).SetError(&r).SetResult(&r).SetBody(b).SetHeaders(map[string]string{
		"x-client-id": webClientID,
		"x-device-id": deviceID,
	}).Post(userURL("/v1/shield/captcha/init"))
	if err != nil {
		return "", err
	}

	log.WithContext(ctx).Debugf("get captcha token resp body: %s", resp.String())
	if err = r.Error(); err != nil {
		return "", err
	}

	return r.CaptchaToken, nil
}
