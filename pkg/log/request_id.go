// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package log

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestIDContext set request id to context. The requestIDHook will read it.
func RequestIDContext(ctx context.Context, id string) context.Context {
	if id == "" {
		id = strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	return context.WithValue(ctx, requestIDHook{}, id)
}

type requestIDHook struct{}

// Fire implements the logrus.Hook interface.
func (hook *requestIDHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}
	if entry.Data == nil {
		entry.Data = make(logrus.Fields)
	}
	id, ok := entry.Context.Value(requestIDHook{}).(string)
	if ok && id != "" {
		entry.Data["request_id"] = id
	}
	return nil
}

// Levels implements the logrus.Hook interface.
func (hook *requestIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
