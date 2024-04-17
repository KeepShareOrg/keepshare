// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/KeepShareOrg/keepshare/pkg/log"
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

// changeHostPassword change host password
func changeHostPassword(c *gin.Context) {
	var r struct {
		RememberMe  bool   `json:"remember_me"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BindJSON(&r); err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	hostName := c.DefaultQuery("host", config.DefaultHost())
	host := hosts.Get(hostName)

	ctx := c.Request.Context()
	userID := c.GetString(constant.UserID)

	taskID, err := host.ChangeMasterAccountPassword(ctx, userID, r.NewPassword, r.RememberMe)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
}

// getChangePasswordTaskInfo get change password task info
func getChangePasswordTaskInfo(c *gin.Context) {
	ctx := c.Request.Context()
	taskID := c.Query("id")

	if taskID == "" {
		mdw.RespInternal(c, "task id is required")
		return
	}

	status, err := config.Redis().Get(ctx, taskID).Result()
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// confirmPassword confirm password
func confirmPassword(c *gin.Context) {
	var r struct {
		Password     string `json:"password"`
		SavePassword bool   `json:"save_password"`
	}

	if err := c.BindJSON(&r); err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	hostName := c.DefaultQuery("host", config.DefaultHost())
	host := hosts.Get(hostName)

	ksUserID := c.GetString(constant.UserID)
	log.Infof("confirming password for keep share user %s %v", ksUserID, r.Password)
	ctx := c.Request.Context()

	err := host.ConfirmMasterAccountPassword(ctx, ksUserID, r.Password, r.SavePassword)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
