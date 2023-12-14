// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package log

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type contextData struct {
	lock         *sync.RWMutex
	requestID    string
	requestStart time.Time
	fields       Fields
}

// DataContextOptions is the options to DataContext.
type DataContextOptions struct {
	RequestID string
	Fields    Fields
}

// DataContext set data to context. The dataContextHook will read it.
// If an empty requestID is passed in, a random one will be generated.
func DataContext(ctx context.Context, opts ...DataContextOptions) context.Context {
	var requestID string
	if len(opts) > 0 && opts[0].RequestID != "" {
		requestID = opts[0].RequestID
	} else {
		requestID = NewRequestID()
	}

	fields := make(Fields)
	if len(opts) > 0 && len(opts[0].Fields) > 0 {
		for k, v := range opts[0].Fields {
			fields[k] = v
		}
	}

	return context.WithValue(ctx, dataContextHook{}, &contextData{
		lock:         new(sync.RWMutex),
		requestID:    requestID,
		requestStart: time.Now(),
		fields:       fields,
	})
}

// ContextWithFields persist fields into context if the context contains dataContextHook.
func ContextWithFields(ctx context.Context, fields Fields) {
	if len(fields) == 0 {
		return
	}

	data, ok := ctx.Value(dataContextHook{}).(*contextData)
	if !ok || data == nil {
		return
	}

	data.lock.RLock()
	defer data.lock.RUnlock()
	for k, v := range fields {
		data.fields[k] = v
	}
}

// NewRequestID returns a random request id.
func NewRequestID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

type dataContextHook struct{}

// Fire implements the logrus.Hook interface.
func (hook *dataContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}

	data, ok := entry.Context.Value(dataContextHook{}).(*contextData)
	if !ok || data == nil {
		return nil
	}

	if entry.Data == nil {
		entry.Data = make(logrus.Fields)
	}

	data.lock.RLock()
	defer data.lock.RUnlock()
	for k, v := range data.fields {
		if _, ok := entry.Data[k]; !ok {
			entry.Data[k] = v // lowest priority
		}
	}

	// highest priority
	entry.Data["request_id"] = data.requestID
	entry.Data["request_ns"] = time.Since(data.requestStart).Nanoseconds()

	return nil
}

// Levels implements the logrus.Hook interface.
func (hook *dataContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
