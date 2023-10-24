// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"net/http"

	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/server/auth"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
)

func signUp(c *gin.Context) {
	type request struct {
		Email        string `json:"email"`
		PasswordHash string `json:"password_hash"`
		CaptchaToken string `json:"captcha_token"`
	}
	req := new(request)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	if req.Email == "" || req.PasswordHash == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "email and password_hash are required")))
		return
	}

	if req.CaptchaToken == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "captcha_token is required")))
		return
	}

	if !VerifyRecaptchaToken(req.CaptchaToken) {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid recaptcha token")))
		return
	}

	user := &model.User{
		ID:           auth.NewID(),
		Channel:      auth.NewChannelId(),
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}

	err := query.User.WithContext(c.Request.Context()).Create(user)

	if err != nil {
		if gormutil.IsDuplicateError(err) {
			c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "duplicate_user", i18n.WithDataMap("error", err.Error())))
			return
		}
		mdw.RespInternal(c, err.Error())
		return
	}

	tokens, err := mdw.GenerateTokens(&mdw.Token{
		UserId:    user.ID,
		ChannelId: user.Channel,
		Email:     user.Email,
		Username:  user.Name,
	})
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, Map{
		"ok":            true,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func signIn(c *gin.Context) {
	type request struct {
		Email        string `json:"email"`
		PasswordHash string `json:"password_hash"`
		CaptchaToken string `json:"captcha_token"`
	}
	req := new(request)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	if req.Email == "" || req.PasswordHash == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "email and password_hash are required")))
		return
	}

	if req.CaptchaToken == "" {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "captcha_token is required")))
		return
	}

	if !VerifyRecaptchaToken(req.CaptchaToken) {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid recaptcha token")))
		return
	}

	user, err := query.User.WithContext(c.Request.Context()).Where(query.User.Email.Eq(req.Email)).Take()
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	if user == nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "account_verify_failed"))
		return
	}

	if req.PasswordHash != user.PasswordHash {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "account_verify_failed"))
		return
	}

	tokens, err := mdw.GenerateTokens(&mdw.Token{
		UserId:    user.ID,
		ChannelId: user.Channel,
		Email:     user.Email,
		Username:  user.Name,
	})
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, Map{
		"ok":            true,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func signOut(c *gin.Context) {

}

func refreshToken(c *gin.Context) {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}
	req := new(request)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	tm, err := mdw.NewTokenManager()
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}
	info, err := tm.ValidateToken(req.RefreshToken)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	newToken := &mdw.Token{
		UserId:    info.UserId,
		ChannelId: info.ChannelId,
		Email:     info.Email,
		Username:  info.Username,
	}
	accessToken, err := tm.GenerateAccessToken(newToken)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}
	refreshToken, err := tm.GenerateRefreshToken(newToken)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, Map{
		"ok":            true,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func getUserInfo(c *gin.Context) {
	userID := c.GetString(constant.UserID)
	channelID := c.GetString(constant.ChannelID)
	email := c.GetString(constant.Email)
	username := c.GetString(constant.Username)

	c.JSON(http.StatusOK, Map{
		"ok":         true,
		"user_id":    userID,
		"channel_id": channelID,
		"email":      email,
		"username":   username,
	})
}
