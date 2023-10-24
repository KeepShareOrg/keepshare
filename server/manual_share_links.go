// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

func createSharedLinks(c *gin.Context) {
	type request struct {
		Links []string `json:"links"`
	}

	req := new(request)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params"))
		return
	}

	hostName := util.FirstNotEmpty(c.Query("host"), config.DefaultHost())
	host := hosts.Get(hostName)
	if host == nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_host", i18n.WithDataMap("host", hostName)))
		return
	}

	if len(req.Links) == 0 {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "links_is_empty"))
		return
	}

	links := make([]string, 0)
	for _, link := range req.Links {
		if simple, _, ok := validateLink(link); ok {
			links = append(links, simple)
		}
	}

	userID := c.GetString(constant.UserID)
	tasks := make([]*model.SharedLink, 0)
	for _, link := range links {
		t := &model.SharedLink{
			UserID:             userID,
			State:              share.StatusPending.String(),
			Host:               hostName,
			CreatedBy:          share.LinkToShare,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
			OriginalLinkHash:   lk.Hash(link),
			HostSharedLinkHash: "",
			OriginalLink:       link,
			HostSharedLink:     "",
		}
		tasks = append(tasks, t)
	}

	ctx := c.Request.Context()
	if err := query.SharedLink.WithContext(ctx).
		Clauses(clause.Insert{Modifier: "IGNORE"}).
		Create(tasks...); err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"links":   tasks,
	})
}
