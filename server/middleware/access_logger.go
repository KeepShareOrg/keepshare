// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"net"
	"os"
	"regexp"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var accessLogger gin.HandlerFunc

// SkipAccessLog is a flag to skip access log
var SkipAccessLog = "_skip_access_log_"

// AccessLogger a Logger middleware that will write the access logs to stdout or file.
func AccessLogger(pathPatterns ...*regexp.Regexp) gin.HandlerFunc {
	if accessLogger != nil {
		return accessLogger
	}

	var logger *logrus.Logger
	if o := config.AccessLogOutput(); o == "" || o == config.LogOutput() {
		logger = log.Log()
	} else {
		logger = log.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(log.JSONLogFormatter)
		logger.SetOutput(log.Writer(o))
	}

	server, _ := os.Hostname()

	accessLogger = func(c *gin.Context) {
		ctx := log.DataContext(c.Request.Context(), log.DataContextOptions{
			Fields: log.Fields{"src": "request"},
		})
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		path := c.Request.URL.RawPath
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		if c.GetBool(SkipAccessLog) {
			return
		}
		if len(pathPatterns) > 0 && !matchAny(path, pathPatterns) {
			return
		}

		host, _, _ := net.SplitHostPort(c.Request.Host)
		data := map[string]any{
			"ip":         c.ClientIP(),
			"latency_ms": float64(time.Since(start).Microseconds()) / 1000,
			"method":     c.Request.Method,
			"proto":      c.Request.Proto,
			"status":     c.Writer.Status(),
			"resp_size":  c.Writer.Size(),
			"path":       path,
			"query":      query,
			"refer":      c.Request.Header.Get("Referer"),
			"user_agent": c.Request.Header.Get("User-Agent"),
			"node":       host,
			"server":     server,
		}

		for _, k := range []string{constant.UserID, constant.Error} {
			if v, exists := c.Get(k); exists {
				data[k] = v
			}
		}

		logger.WithContext(ctx).WithFields(data).Info(c.GetString(constant.Message))
	}

	return accessLogger
}

func matchAny(s string, rs []*regexp.Regexp) bool {
	for _, r := range rs {
		if r.MatchString(s) {
			return true
		}
	}
	return false
}
