// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const blacklistSubmitLimit = 1000

func getBlackList(c *gin.Context) {
	type blackListLink struct {
		OriginalLink    string    `json:"original_link"`
		KeepSharingLink string    `json:"keep_sharing_link"`
		CreatedAt       time.Time `json:"created_at"`
	}
	var resp struct {
		Total    int              `json:"total"`
		PageSize int              `json:"page_size"`
		List     []*blackListLink `json:"list"`
	}

	ctx := c.Request.Context()
	t := query.Blacklist
	userID := c.GetString(constant.UserID)
	channel := c.GetString(constant.ChannelID)
	page, _ := strconv.Atoi(c.Query("page_index"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit > blacklistSubmitLimit || limit <= 0 {
		limit = blacklistSubmitLimit
	}

	ret, count, err := t.WithContext(ctx).Where(t.UserID.Eq(userID)).Order(t.CreatedAt.Desc()).FindByPage(page*limit, limit)
	if gormutil.IsNotFoundError(err) {
		c.JSON(http.StatusOK, resp)
		return
	}
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	resp.Total = int(count)
	resp.PageSize = len(ret)
	for _, r := range ret {
		resp.List = append(resp.List, &blackListLink{
			OriginalLink:    r.OriginalLink,
			KeepSharingLink: makeKeepSharingLink(channel, r.OriginalLink),
			CreatedAt:       r.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, resp)
}

func addToBlackList(c *gin.Context) {
	var req struct {
		Links  []string `json:"links"`
		Delete bool     `json:"delete"`
	}
	var resp struct {
		OKCount    int      `json:"ok_count"`
		ErrorCount int      `json:"error_count"`
		ErrorLinks []string `json:"error_links"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_request", i18n.WithDataMap("error", err.Error())))
		return
	}
	if len(req.Links) == 0 {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "no links")))
		return
	}
	if len(req.Links) > blacklistSubmitLimit {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "submit_too_many_links", i18n.WithDataMap("count", strconv.Itoa(blacklistSubmitLimit))))
		return
	}

	userID := c.GetString(constant.UserID)

	original, invalid := getOriginalLinks(req.Links)
	resp.OKCount = len(original)
	resp.ErrorCount = len(invalid)
	resp.ErrorLinks = invalid

	if resp.OKCount == 0 {
		c.JSON(http.StatusOK, resp)
		return
	}

	rows := make([]*model.Blacklist, 0, len(original))
	hashes := make([]string, 0, len(original))
	now := time.Now()
	for _, v := range original {
		r := &model.Blacklist{
			UserID:           userID,
			OriginalLinkHash: lk.Hash(v),
			OriginalLink:     v,
			CreatedAt:        now,
		}
		rows = append(rows, r)
		hashes = append(hashes, r.OriginalLinkHash)
	}

	ctx := c.Request.Context()
	if err := query.Blacklist.WithContext(ctx).Clauses(
		clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{query.Blacklist.CreatedAt.ColumnName().String()}),
		},
	).Create(rows...); err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	if req.Delete {
		for _, h := range hosts.GetAll() {
			if _, err := deleteSharedLinksWithHost(ctx, userID, h, original); err != nil {
				mdw.RespInternal(c, err.Error())
				return
			}
		}
	} else {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			t := query.SharedLink
			_, _ = t.WithContext(ctx).Where(
				t.UserID.Eq(userID),
				t.State.Neq(share.StatusBlocked.String()),
				t.OriginalLinkHash.In(hashes...),
			).Select(t.State, t.UpdatedAt).Updates(&model.SharedLink{
				State:     share.StatusBlocked.String(),
				UpdatedAt: time.Now(),
			})
		}()
	}

	c.JSON(http.StatusOK, resp)
}

func removeFromBlackList(c *gin.Context) {
	var req struct {
		Links []string `json:"links"`
	}
	var resp struct {
		OKCount    int      `json:"ok_count"`
		ErrorCount int      `json:"error_count"`
		ErrorLinks []string `json:"error_links"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_request", i18n.WithDataMap("error", err.Error())))
		return
	}
	if len(req.Links) == 0 {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if len(req.Links) > blacklistSubmitLimit {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "submit_too_many_links", i18n.WithDataMap("count", strconv.Itoa(blacklistSubmitLimit))))
		return
	}

	userID := c.GetString(constant.UserID)

	original, invalid := getOriginalLinks(req.Links)
	resp.OKCount = len(original)
	resp.ErrorCount = len(invalid)
	resp.ErrorLinks = invalid

	if resp.ErrorCount > 0 {
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	if resp.OKCount == 0 {
		c.JSON(http.StatusOK, resp)
		return
	}

	hashes := make([]string, 0, len(original))
	for _, v := range original {
		hashes = append(hashes, lk.Hash(v))
	}

	ctx := c.Request.Context()
	t := query.Blacklist
	_, err := t.WithContext(ctx).Where(t.UserID.Eq(userID), t.OriginalLinkHash.In(hashes...)).Delete()
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	go func() {
		t := query.SharedLink
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// records with host shared link update status to Complete.
		// records without host shared link update to Processing.
		w := fmt.Sprintf("`%s` = ? AND `%s` = ? AND `%s` IN (?)",
			t.UserID.ColumnName().String(),
			t.State.ColumnName().String(),
			t.OriginalLinkHash.ColumnName().String(),
		)
		ifExpr := fmt.Sprintf("IF(`%s` = '', ?, ?)", t.HostSharedLinkHash.ColumnName().String())
		config.MySQL().WithContext(ctx).
			Table(model.TableNameSharedLink).
			Where(w, userID, share.StatusBlocked.String(), hashes).Updates(map[string]any{
			t.UpdatedAt.ColumnName().String(): gorm.Expr(
				ifExpr,
				gorm.Expr(t.CreatedAt.ColumnName().String()), // to trigger asyncTaskCheckBackground
				time.Now(),
			),
			t.State.ColumnName().String(): gorm.Expr(
				ifExpr,
				share.StatusPending.String(),
				share.StatusOK.String(),
			)})

		// update all records status to StatusCreated.
		//_, _ = t.WithContext(ctx).Where(
		//	t.UserID.Eq(userID),
		//	t.State.Eq(share.StatusBlocked.String()),
		//	t.OriginalLinkHash.In(hashes...),
		//).UpdateColumns(map[string]any{
		//	t.UpdatedAt.ColumnName().String(): t.CreatedAt, // to trigger asyncTaskCheckBackground
		//	t.State.ColumnName().String():     share.StatusCreated.String(),
		//})
	}()

	c.JSON(http.StatusOK, resp)
}
