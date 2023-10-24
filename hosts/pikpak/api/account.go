// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/pkg/mail"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	log "github.com/sirupsen/logrus"
)

// UserInfo is the info of a user.
type UserInfo struct {
	UserID       string `json:"user_id"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// SignUp sign up a new account by email.
func (api *API) SignUp(ctx context.Context, email string, timeout time.Duration) (user *UserInfo, err error) {
	if email == "" {
		email = api.randomEmail()
	}

	start := time.Now()
	defer func() {
		l := log.WithFields(log.Fields{
			"email":      email,
			"latency_ms": int(time.Since(start) / time.Millisecond),
		})
		if err != nil {
			l.WithError(err).Errorf("signup err")
		} else {
			l.Infof("signup user_id: %s", user.UserID)
		}
	}()

	device := randomDevice()

	captchaToken, err := api.signupCaptcha(ctx, email, device)
	if err != nil {
		return nil, err
	}

	verificationID, err := api.signupSendEmail(ctx, email, captchaToken, device)
	if err != nil {
		return nil, err
	}

	var code string
	var found bool
	rounds := int(timeout / time.Second)
	if rounds <= 0 {
		rounds = 5
	}
	for round := 1; round <= rounds; round++ {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		if round == 1 {
			time.Sleep(time.Second / 2)
		} else {
			time.Sleep(time.Second)
		}
		if err = ctx.Err(); err != nil {
			return nil, err
		}

		code, found, err = api.signupGetCode(ctx, email, start)
		if err != nil {
			return nil, err
		}
		if found {
			break
		}
	}

	if code == "" {
		return nil, fmt.Errorf("get signup code failed in %d ms", time.Since(start)/time.Millisecond)
	}

	token, err := api.signupVerifyCode(ctx, code, verificationID, device)
	if err != nil {
		return nil, err
	}

	return api.signup(ctx, email, code, token, randomPassword(), device)
}

func (api *API) signupCaptcha(ctx context.Context, email string, deviceID string) (verificationID string, err error) {
	var r struct {
		*RespErr
		CaptchaToken string `json:"captcha_token"`
	}

	b := util.ToJSON(map[string]any{
		"action":    "POST:/v1/auth/verification",
		"client_id": webClientID,
		"device_id": deviceID,
		"meta": map[string]string{
			"email": email,
		},
	})
	body, err := resCli.R().
		SetContext(ctx).
		SetError(&r).
		SetResult(&r).
		SetBody(b).
		SetHeader("x-client-id", webClientID).
		SetHeader("x-device-id", deviceID).
		Post(userURL("/v1/shield/captcha/init"))

	if err != nil {
		return "", fmt.Errorf("get captcha token err: %w", err)
	}

	log.Debugf("get captcha token resp body: %s", body.Body())

	if err = r.Error(); err != nil {
		return "", fmt.Errorf("get captcha token err: %w", err)
	}

	return r.CaptchaToken, nil
}

func (api *API) signupSendEmail(ctx context.Context, email, captcha, deviceID string) (verificationID string, err error) {
	var r struct {
		*RespErr
		VerificationID string `json:"verification_id"`
	}

	b := util.ToJSON(map[string]any{
		"email":     email,
		"target":    "ANY",
		"usage":     "REGISTER",
		"locale":    "en-US",
		"client_id": webClientID,
	})
	body, err := resCli.R().
		SetContext(ctx).
		SetError(&r).
		SetResult(&r).
		SetHeader("x-captcha-token", captcha).
		SetHeader("x-client-id", webClientID).
		SetHeader("x-device-id", deviceID).
		SetBody(b).
		Post(userURL("/v1/auth/verification"))

	if err != nil {
		return "", fmt.Errorf("send signup email err: %w", err)
	}

	log.Debugf("send signup email resp body: %s", body.Body())

	if err = r.Error(); err != nil {
		return "", fmt.Errorf("send signup email err: %w", err)
	}

	return r.VerificationID, nil
}

var signupCodeRegexp = regexp.MustCompile(`[0-9]{6}`)

func (api *API) signupGetCode(ctx context.Context, email string, sentTime time.Time) (code string, found bool, err error) {
	headers, err := api.Mailer.List(ctx, email)
	if err != nil {
		return "", false, fmt.Errorf("list mail err: %w", err)
	}

	const from = "<noreply@accounts.mypikpak.com>"
	var header *mail.Header
	// The newest emails are at the end of the list.
	for i := len(headers) - 1; i >= 0; i-- {
		h := headers[i]
		if h.Date.Before(sentTime) {
			continue
		}
		if strings.Contains(h.From, from) {
			header = headers[i]
			break
		}
	}

	l := log.WithField("email", email)
	if header == nil {
		l.Debugf("not found the email from %s and sent after %s", from, sentTime.Format(time.RFC3339))
		return "", false, nil
	}

	body, err := api.Mailer.Get(ctx, email, header.ID)
	if err != nil {
		return "", false, fmt.Errorf("get mail body err: %w", err)
	}
	code = signupCodeRegexp.FindString(body.Text)
	if code != "" {
		// clear mail records after find code.
		api.Mailer.Clear(ctx, email)
		l.Debugf("signup code: %s", code)
		return code, true, nil
	}
	l.Debugf("email found but code not found")
	return "", false, nil
}

func (api *API) signupVerifyCode(ctx context.Context, code string, verificationID, deviceID string) (token string, err error) {
	var r struct {
		*RespErr
		VerificationToken string `json:"verification_token"`
	}

	b := util.ToJSON(map[string]any{
		"verification_id":   verificationID,
		"verification_code": code,
		"client_id":         webClientID,
	})
	body, err := resCli.R().
		SetContext(ctx).
		SetError(&r).
		SetResult(&r).
		SetBody(b).
		SetHeader("x-client-id", webClientID).
		SetHeader("x-device-id", deviceID).
		Post(userURL("/v1/auth/verification/verify"))

	if err != nil {
		return "", fmt.Errorf("verify signup token err: %w", err)
	}

	log.Debugf("verify signup token resp body: %s", body.Body())

	if err = r.Error(); err != nil {
		return "", fmt.Errorf("verify signup token err: %w", err)
	}

	return r.VerificationToken, nil
}

func (api *API) signup(ctx context.Context, email, code, token, password, deviceID string) (user *UserInfo, err error) {
	var r struct {
		*RespErr
		UserID       string `json:"sub"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	b := util.ToJSON(map[string]any{
		"email":              email,
		"verification_code":  code,
		"verification_token": token,
		"password":           password,
		"client_id":          webClientID,
	})
	body, err := resCli.R().
		SetContext(ctx).
		SetError(&r).
		SetResult(&r).
		SetBody(b).
		SetHeaders(map[string]string{
			"x-device-sign": fmt.Sprintf("wdi10.%sxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", deviceID),
			"x-client-id":   webClientID,
			"x-device-id":   deviceID,

			"x-provider-name":    "NONE",
			"x-sdk-version":      "6.0.0",
			"accept-language":    "en-US",
			"x-os-version":       "Win32",
			"x-net-work-type":    "NONE",
			"sec-ch-ua-platform": `"Windows"`,
			"x-platform-version": "1",
			"x-protocol-version": "301",
			"x-client-version":   "1.0.0",
			"content-type":       "application/json",
			"Referer":            "https://mypikpak.com/",
			"x-device-model":     "chrome/115.0.0.0",
			"x-device-name":      "PC-Chrome",
		}).
		Post(userURL("/v1/auth/signup"))

	if err != nil {
		return nil, fmt.Errorf("signup err: %w", err)
	}

	log.Debugf("signup resp body: %s", body.Body())

	if err = r.Error(); err != nil {
		return nil, fmt.Errorf("signup err: %w", err)
	}
	if r.UserID == "" {
		return nil, fmt.Errorf("signup got unexpected response: %s", body.Body())
	}

	return &UserInfo{
		UserID:       r.UserID,
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresIn:    r.ExpiresIn,
		Password:     password,
		Email:        email,
	}, nil
}
