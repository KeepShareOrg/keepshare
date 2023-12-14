// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package log

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type requestIDValue struct {
	id    string
	start time.Time
}

// RequestIDContext set request id to context. The requestIDHook will read it.
// If an empty id is passed in, a randomly one will be generated.
func RequestIDContext(ctx context.Context, id ...string) context.Context {
	var s string
	if len(id) == 0 || id[0] == "" {
		s = strings.ReplaceAll(uuid.NewString(), "-", "")
	} else {
		s = id[0]
	}
	return context.WithValue(ctx, requestIDHook{}, requestIDValue{s, time.Now()})
}

type requestIDHook struct{}

// Fire implements the logrus.Hook interface.
func (hook *requestIDHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}

	v, ok := entry.Context.Value(requestIDHook{}).(requestIDValue)
	if !ok {
		return nil
	}

	if entry.Data == nil {
		entry.Data = make(logrus.Fields)
	}
	entry.Data["request_id"] = v.id
	entry.Data["request_ns"] = time.Since(v.start).Nanoseconds()
	return nil
}

// Levels implements the logrus.Hook interface.
func (hook *requestIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
