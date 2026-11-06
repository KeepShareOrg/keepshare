// Copyright 2026 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func isLandingPage(path string) bool {
	return path == "/console/shared/status" || path == "/console/shared/wsl-status"
}

func recordLandingPageVisit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet || !isLandingPage(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Each document load must reach the server; status API polls never count.
		c.Header("Cache-Control", "no-store")
		c.Next()
		if c.Writer.Status() != http.StatusOK {
			return
		}

		linkID, err := strconv.ParseInt(c.Query("id"), 10, 64)
		if err != nil || linkID <= 0 {
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if err := recordLinkVisit(ctx, db, linkID, c.ClientIP()); err != nil {
			log.WithContext(ctx).WithField("shared_link_id", linkID).WithField("error", err).
				Error("record landing page visit failed")
		}
	}
}
