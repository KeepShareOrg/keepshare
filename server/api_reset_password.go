// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

func sendVerificationCode(c *gin.Context) {
	type Req struct {
		Email string `json:"email"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid email")))
		return
	}

	ctx := c.Request.Context()

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

	if _, err := config.Redis().
		SetEx(ctx, verificationToken, verificationCode, verificationCodeExpire).
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

	if err := emailClient.NewMessage("Verify Email").
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
	// TODO: Limit the number of retries
	type Req struct {
		VerificationCode  string `json:"verification_code"`
		VerificationToken string `json:"verification_token"`
	}

	var req Req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	if req.VerificationCode == "" || req.VerificationToken == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid params")))
		return
	}

	ctx := c.Request.Context()
	verificationCode, err := config.Redis().Get(ctx, req.VerificationToken).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid verification code")))
		return
	}

	if req.VerificationCode != verificationCode {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid verification code")))
		return
	}

	config.Redis().Del(ctx, req.VerificationToken)

	c.JSON(200, gin.H{"message": "ok"})
}
