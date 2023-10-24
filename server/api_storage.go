// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
)

const notStoredDaysMax = 36500 // to avoid exceeding the minimum time.

func storageStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Host            string  `json:"host"`
		StoredCountLt   []int32 `json:"stored_count_lt"`
		NotStoredDaysGt []int32 `json:"not_stored_days_gt"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	total := len(req.StoredCountLt) + len(req.NotStoredDaysGt)
	if total == 0 {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "at least one condition is required")))
		return
	}
	if total > 10 {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "too many conditions")))
		return
	}

	hostName := util.FirstNotEmpty(req.Host, config.DefaultHost())
	userID := c.GetString(constant.UserID)

	colStored := query.SharedLink.Stored.ColumnName().String()
	colLastStoredAt := query.SharedLink.LastStoredAt.ColumnName().String()
	colSize := query.SharedLink.Size.ColumnName().String()

	//storedCountLt := []int64{1, 10, 100, 1000}
	//notStoredDaysGt := []int64{60, 180, 365}

	for i, v := range req.NotStoredDaysGt {
		if v > notStoredDaysMax {
			req.NotStoredDaysGt[i] = notStoredDaysMax
		}
	}

	var selects []string
	var results []any

	for _, v := range req.StoredCountLt {
		selects = append(selects, fmt.Sprintf("SUM(IF(`%s` < %d, 1, 0))", colStored, v))
		selects = append(selects, fmt.Sprintf("SUM(IF(`%s` < %d, `%s`, 0))", colStored, v, colSize))
		results = append(results, &sql.NullInt64{}, &sql.NullInt64{})
	}
	for _, v := range req.NotStoredDaysGt {
		t := time.Now().Add(time.Duration(-1*v) * 24 * time.Hour)
		selects = append(selects, fmt.Sprintf("SUM(IF(`%s` < '%s 00:00:00', 1, 0))", colLastStoredAt, t.Format(time.DateOnly)))
		selects = append(selects, fmt.Sprintf("SUM(IF(`%s` < '%s 00:00:00', `%s`, 0))", colLastStoredAt, t.Format(time.DateOnly), colSize))
		results = append(results, &sql.NullInt64{}, &sql.NullInt64{})
	}

	db := config.MySQL().WithContext(ctx).Table(query.SharedLink.TableName())
	err := db.
		Select(strings.Join(selects, ", ")).
		Where("`user_id` = ? AND `host` = ?", userID, hostName).
		Row().
		Scan(results...)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	type data struct {
		Number     int32 `json:"number"`
		TotalCount int64 `json:"total_count"`
		TotalSize  int64 `json:"total_size"`
	}
	var resp struct {
		StoredCountLt   []data `json:"stored_count_lt"`
		NotStoredDaysGt []data `json:"not_stored_days_gt"`
	}

	var index int
	for _, v := range req.StoredCountLt {
		resp.StoredCountLt = append(resp.StoredCountLt, data{
			Number:     v,
			TotalCount: results[index].(*sql.NullInt64).Int64,
			TotalSize:  results[index+1].(*sql.NullInt64).Int64,
		})
		index += 2
	}
	for _, v := range req.NotStoredDaysGt {
		resp.NotStoredDaysGt = append(resp.NotStoredDaysGt, data{
			Number:     v,
			TotalCount: results[index].(*sql.NullInt64).Int64,
			TotalSize:  results[index+1].(*sql.NullInt64).Int64,
		})
		index += 2
	}

	c.JSON(http.StatusOK, resp)
}

func storageRelease(c *gin.Context) {
	var req struct {
		Host            string `json:"host"`
		StoredCountLt   int32  `json:"stored_count_lt"`
		NotStoredDaysGt int32  `json:"not_stored_days_gt"`
		OnlyForPremium  bool   `json:"only_for_premium"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	hostName := util.FirstNotEmpty(req.Host, config.DefaultHost())
	if req.NotStoredDaysGt > notStoredDaysMax {
		req.NotStoredDaysGt = notStoredDaysMax // Avoid exceeding the minimum time.
	}

	host := hosts.Get(hostName)
	if host == nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_host", i18n.WithDataMap("host", hostName)))
		return
	}

	ctx := c.Request.Context()
	userID := c.GetString(constant.UserID)
	t := query.SharedLink
	lastTime := time.Now().Add(time.Duration(-1*req.NotStoredDaysGt) * 24 * time.Hour)

	stmt := t.WithContext(ctx).Where(t.UserID.Eq(userID), t.Host.Eq(hostName))

	switch {
	case req.StoredCountLt <= 0 && req.NotStoredDaysGt <= 0:
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "one of stored_count_lt or not_stored_days_gt is required")))
		return

	case req.StoredCountLt > 0 && req.NotStoredDaysGt > 0:
		stmt = stmt.Where(t.Where(t.Stored.Lt(req.StoredCountLt)).Or(t.LastStoredAt.Lt(lastTime)))

	case req.StoredCountLt > 0:
		stmt = stmt.Where(t.Stored.Lt(req.StoredCountLt))

	default:
		stmt = stmt.Where(t.LastStoredAt.Lt(lastTime))
	}

	rows, err := stmt.Find()
	if err != nil && !gormutil.IsNotFoundError(err) {
		mdw.RespInternal(c, err.Error())
		return
	}

	if req.OnlyForPremium {
		// TODO
	}

	var resp struct {
		RowsAffected int `json:"rows_affected"`
	}

	if len(rows) == 0 {
		c.JSON(http.StatusOK, resp)
		return
	}

	originalLinks := make([]string, 0, len(rows))
	autoIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		originalLinks = append(originalLinks, row.OriginalLink)
		autoIDs = append(autoIDs, row.AutoID)
	}
	if err := host.Delete(ctx, userID, originalLinks); err != nil {
		mdw.RespInternal(c, fmt.Sprintf("delete host shared link err: %v", err))
		return
	}
	ret, err := t.WithContext(ctx).Where(t.AutoID.In(autoIDs...)).Delete()
	if err != nil {
		mdw.RespInternal(c, fmt.Sprintf("delete shared records err: %v", err))
		return
	}

	resp.RowsAffected = int(ret.RowsAffected)
	c.JSON(http.StatusOK, resp)
}
