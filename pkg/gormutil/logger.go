// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gormutil

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

// GormLogger create a grom logger interface.
func GormLogger(level string) logger.Interface {
	var l logger.LogLevel
	switch strings.ToLower(level) {
	case "trace", "debug", "info":
		l = logger.Info
	case "warn", "warning":
		l = logger.Warn
	case "error":
		l = logger.Error
	default:
		l = logger.Silent
	}
	return &gormLogger{
		level: l,
		log:   log.StandardLogger(),
	}
}

type gormLogger struct {
	level logger.LogLevel
	log   *log.Logger
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l *gormLogger) Info(_ context.Context, format string, v ...interface{}) {
	if l.level >= logger.Info {
		l.log.Infof(format, v...)
	}
}

func (l *gormLogger) Warn(_ context.Context, format string, v ...interface{}) {
	if l.level >= logger.Warn {
		l.log.Warnf(format, v...)
	}
}

func (l *gormLogger) Error(_ context.Context, format string, v ...interface{}) {
	if l.level >= logger.Error {
		l.log.Errorf(format, v...)
	}
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= logger.Silent {
		return
	}
	if contextIgnoreTrace(ctx) {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	f := map[string]interface{}{
		"sql":        sql,
		"latency_ms": float64(elapsed.Microseconds()) / 1000,
		"rows":       rows,
	}
	if err != nil && !IsNotFoundError(err) {
		f["error"] = err.Error()
		if l.level >= logger.Error {
			l.log.WithFields(f).Error()
		}
		return
	}

	if elapsed > time.Millisecond*200 { // slow SQL
		if l.level >= logger.Warn {
			l.log.WithFields(f).Warn()
		}
		return
	}

	if l.level >= logger.Info {
		// print SELECT with debug level
		// print others with info level
		if len(sql) > 6 && strings.EqualFold(sql[:6], "SELECT") {
			l.log.WithFields(f).Debug()
		} else {
			l.log.WithFields(f).Info()
		}
	}
}

type noTraceContextKey struct{}

// IgnoreTraceContext returns a context indicates that no trace log is required.
func IgnoreTraceContext(parent context.Context) context.Context {
	key := noTraceContextKey{}
	return context.WithValue(parent, key, key)
}

func contextIgnoreTrace(ctx context.Context) bool {
	return ctx.Value(noTraceContextKey{}) != nil
}
