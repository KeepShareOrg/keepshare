// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	debugToken         string
	debugTokenInitOnce sync.Once
)

// Auth authenticate users.
func Auth(c *gin.Context) {
	debugTokenInitOnce.Do(func() {
		debugToken = viper.GetString("debug_token")
	})

	tokenStr, err := parseTokenFromHeader(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrResp(c, "invalid_token", i18n.WithDataMap("error", err.Error())))
		return
	}

	if debugToken != "" && debugToken == tokenStr {
		c.Set(constant.UserID, "debug")
		c.Set(constant.ChannelID, "00000000")
		c.Set(constant.Email, "debug@local")
		c.Set(constant.Username, "debug")
		return
	}

	tm, err := NewTokenManager()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	token, err := tm.ValidateToken(tokenStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrResp(c, "invalid_token", i18n.WithDataMap("error", err.Error())))
		return
	}
	c.Set(constant.UserID, token.UserId)
	c.Set(constant.ChannelID, token.ChannelId)
	c.Set(constant.Email, token.Email)
	c.Set(constant.Username, token.Username)

	c.Next()
}

func parseTokenFromHeader(req *http.Request) (string, error) {
	h := req.Header.Get("Authorization")
	temp := strings.Split(strings.TrimSpace(h), " ")
	if len(temp) != 2 {
		return "", errors.New("invalid token")
	}
	return strings.TrimSpace(temp[1]), nil
}
