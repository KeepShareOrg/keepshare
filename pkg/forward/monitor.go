// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package forward

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	baseDelay   = 1 * time.Second
	maxDelay    = 30 * time.Second
	monitorPath = "/api/v1/monitor/messages"
)

// MessageSource yields parsed messages. The concrete implementation is Monitor.
type MessageSource interface {
	Next(ctx context.Context) (Message, bool, error)
}

// Monitor maintains a self-healing WebSocket connection to the mail monitor.
type Monitor struct {
	wsURL string
	conn  *websocket.Conn
	delay time.Duration
}

// NewMonitor returns a Monitor. It does not yet establish a connection.
func NewMonitor(wsURL string) (*Monitor, error) {
	if wsURL == "" {
		return nil, errors.New("monitor: wsURL is required")
	}
	return &Monitor{wsURL: wsURL, delay: baseDelay}, nil
}

// resolveMonitorURL picks the explicit URL or derives one from mailServer.
func resolveMonitorURL(explicit, mailServer string) (string, error) {
	if explicit != "" {
		u, err := url.Parse(explicit)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "", fmt.Errorf("monitor: invalid monitor_url %q", explicit)
		}
		return explicit, nil
	}
	if mailServer == "" {
		return "", errors.New("monitor: monitor_url is empty and mail_server is also empty")
	}
	u, err := url.Parse(mailServer)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("monitor: invalid mail_server %q", mailServer)
	}
	var scheme string
	switch {
	case strings.EqualFold(u.Scheme, "https"):
		scheme = "wss://"
	case strings.EqualFold(u.Scheme, "http"):
		scheme = "ws://"
	default:
		return "", fmt.Errorf("monitor: unsupported mail_server scheme %q", u.Scheme)
	}
	return scheme + u.Host + monitorPath, nil
}

// Next returns the next Message. It blocks until a frame arrives, ctx is
// cancelled, or a connection error forces a reconnect (with backoff applied
// internally). On ctx cancellation it returns the ctx error.
func (m *Monitor) Next(ctx context.Context) (Message, bool, error) {
	for {
		if err := ctx.Err(); err != nil {
			return Message{}, false, err
		}
		if m.conn == nil {
			if err := m.dial(ctx); err != nil {
				m.backoffSleep(ctx)
				continue
			}
		}

		msgType, data, err := m.conn.ReadMessage()
		if err != nil {
			m.closeConn()
			continue
		}
		if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
			continue
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			// drop malformed frames; the caller logs accepted/dropped messages.
			continue
		}
		return msg, true, nil
	}
}

// dial establishes a new WS connection.
func (m *Monitor) dial(ctx context.Context) error {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, m.wsURL, nil)
	if err != nil {
		return err
	}
	m.conn = conn
	m.delay = baseDelay
	return nil
}

// backoffSleep doubles m.delay up to maxDelay, blocking until ctx is done.
func (m *Monitor) backoffSleep(ctx context.Context) {
	d := m.delay
	if d < baseDelay {
		d = baseDelay
	}
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
	m.delay *= 2
	if m.delay > maxDelay {
		m.delay = maxDelay
	}
}

func (m *Monitor) closeConn() {
	if m.conn != nil {
		_ = m.conn.Close()
		m.conn = nil
	}
}
