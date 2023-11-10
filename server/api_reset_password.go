// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"errors"
	"fmt"
	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func sendVerificationCode(c *gin.Context) {
	type Req struct {
		Email  string                      `json:"email"`
		Action constant.VerificationAction `json:"action"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	if req.Email == "" || req.Action == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "email or action is empty")))
		return
	}

	if !isSupportAction(req.Action) {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid action")))
		return
	}

	ctx := c.Request.Context()

	_, err := query.User.WithContext(ctx).Where(query.User.Email.Eq(req.Email)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadGateway, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", "user not found")))
			return
		}
		c.JSON(http.StatusBadGateway, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	verificationCodeExpireConfig := viper.GetString("verify_code_expires")
	if verificationCodeExpireConfig == "" {
		verificationCodeExpireConfig = "10m"
	}
	verificationCodeExpire, err := time.ParseDuration(verificationCodeExpireConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	verificationCodeLength := viper.GetInt("verify_code_length")
	if verificationCodeLength == 0 {
		verificationCodeLength = 6
	}
	verificationCode := GenerateVerificationCode(verificationCodeLength)

	verificationToken := CalcSha265Hash(fmt.Sprintf("%v-%v", req.Email, time.Now().UnixNano()), viper.GetString("verify_code_salt"))

	verificationString := fmt.Sprintf("%v-%v-%v", verificationCode, req.Email, req.Action)
	if _, err := config.Redis().
		SetEx(ctx, verificationToken, verificationString, verificationCodeExpire).
		Result(); err != nil {
		c.JSON(http.StatusInternalServerError, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	emailClient, err := GetEmailClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}
	emailHTMLTemplate := viper.GetString("reset_password_verify_email_html_template")
	emailTextTemplate := viper.GetString("reset_password_verify_email_text_template")

	if err := emailClient.NewMessage("KeepShare - Verify your email").
		AddHtmlContent(fmt.Sprintf(emailHTMLTemplate, verificationCode)).
		AddTextContent(fmt.Sprintf(emailTextTemplate, verificationCode)).
		Send([]string{req.Email}); err != nil {
		c.JSON(http.StatusInternalServerError, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":            "ok",
		"verification_token": verificationToken,
	})
}

func resetPassword(c *gin.Context) {
	type Req struct {
		Email             string                      `json:"email"`
		Action            constant.VerificationAction `json:"action"`
		PasswordHash      string                      `json:"password_hash"`
		VerificationCode  string                      `json:"verification_code"`
		VerificationToken string                      `json:"verification_token"`
	}

	var req Req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	if req.Email == "" || req.PasswordHash == "" || req.Action == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "email, password_hash or action is empty")))
		return
	}

	if !isSupportAction(req.Action) {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid action")))
		return
	}

	if req.VerificationCode == "" || req.VerificationToken == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "verify code or verify token is empty")))
		return
	}

	ctx := c.Request.Context()
	cacheVerificationString, err := config.Redis().Get(ctx, req.VerificationToken).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid verification code")))
		return
	}

	verificationCodeExpireConfig := viper.GetString("verify_code_expires")
	if verificationCodeExpireConfig == "" {
		verificationCodeExpireConfig = "10m"
	}
	verificationRetryCountLimit := viper.GetInt64("verification_retry_count_limit")
	if verificationRetryCountLimit == 0 {
		verificationRetryCountLimit = 3
	}
	verificationCodeExpire, err := time.ParseDuration(verificationCodeExpireConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	verificationString := fmt.Sprintf("%v-%v-%v", req.VerificationCode, req.Email, req.Action)
	if verificationString != cacheVerificationString {
		retryCountKey := fmt.Sprintf("retry_count_%v", req.VerificationToken)
		if count, err := config.Redis().Incr(ctx, retryCountKey).Result(); err == nil {
			if err := config.Redis().ExpireAt(ctx, retryCountKey, time.Now().Add(verificationCodeExpire)).Err(); err != nil {
				log.Errorf("set retry count expired err: %v", err)
			}

			if count > verificationRetryCountLimit {
				config.Redis().Del(ctx, req.VerificationToken)
				c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "Too many retries, please resend the verification code")))
				return
			}
		}

		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "verification failed")))
		return
	}

	config.Redis().Del(ctx, req.VerificationToken)

	if _, err := query.User.WithContext(ctx).
		Where(query.User.Email.Eq(req.Email)).
		Update(query.User.PasswordHash, req.PasswordHash); err != nil {
		c.JSON(http.StatusInternalServerError, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	c.JSON(200, gin.H{"message": "ok"})
}

func isSupportAction(action constant.VerificationAction) bool {
	switch action {
	case constant.VerificationActionResetPassword,
		constant.VerificationActionChangeEmail,
		constant.VerificationActionChangePassword:
		return true
	default:
		return false
	}
}
