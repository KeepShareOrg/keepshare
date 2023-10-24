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

	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(config.JSONLogFormatter)
	log.SetOutput(config.LogWriter(config.AccessLogOutput()))

	server, _ := os.Hostname()

	accessLogger = func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
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

		log.WithFields(data).Info(c.GetString(constant.Message))
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
