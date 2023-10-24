// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"net/http"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/gin-gonic/gin"
)

func getHostInfo(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString(constant.UserID)

	q := c.Request.URL.Query()
	opt := make(map[string]any, len(q))
	for k, v := range q {
		if len(v) > 0 {
			opt[k] = v[0]
		} else {
			opt[k] = nil
		}
	}

	hostName, _ := opt["host"].(string)
	if hostName == "" {
		hostName = config.DefaultHost()
	}
	host := hosts.Get(hostName)
	if host == nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_host", i18n.WithDataMap("host", hostName)))
		return
	}

	resp, err := host.HostInfo(ctx, userID, opt)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, resp)
}

//func releaseStorage(c *gin.Context) {
//	ctx := c.Request.Context()
//	userID := c.GetString(constant.UserID)
//
//	opt := make(map[string]any)
//	if err := c.BindJSON(&opt); err != nil {
//		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_request", i18n.WithDataMap("error", err.Error())))
//		return
//	}
//
//  hostName, _ := opt["host"].(string)
//	if hostName == "" {
//		hostName = config.DefaultHost()
//	}
//	host := hosts.Get(hostName)
//	if host == nil {
//		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_host", i18n.WithDataMap("host", hostName)))
//		return
//	}
//
//	resp, err := host.ReleaseStorage(ctx, userID, opt)
//	if err != nil {
//		mdw.RespInternal(c, err.Error())
//		return
//	}
//
//	c.JSON(http.StatusOK, resp)
//}
