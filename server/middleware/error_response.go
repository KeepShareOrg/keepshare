// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ErrResp response the localized message according to the set key
func ErrResp(c *gin.Context, key string, opts ...i18n.Option) map[string]any {
	lang := c.Request.Header.Get("Accept-Language")
	if lang != "" {
		opts = append(opts, i18n.WithLanguages(lang))
	}

	msg, err := i18n.Get(context.Background(), key, opts...)
	if err != nil {
		log.Errorf("get i18n message for key:%s, language:%s, err:%v", key, lang, err)
	}

	// for access log
	c.Set(constant.Error, key)
	c.Set(constant.Message, msg)

	return map[string]any{constant.Error: key, constant.Message: msg}
}

// RespInternal gin response internal server error with msg.
func RespInternal(c *gin.Context, msg ...string) {
	var opts []i18n.Option
	if len(msg) > 0 {
		opts = []i18n.Option{i18n.WithDataMap("error", strings.Join(msg, ","))}
	}
	c.JSON(http.StatusInternalServerError, ErrResp(c, "internal", opts...))
}
