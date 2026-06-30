// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package forward listens to a real-time mail feed and forwards
// PikPak master-account-bound notifications to the matching keepshare
// user. See docs/superpowers/specs/2026-06-29-master-account-email-forwarder-design.md
// for the design spec.
package forward

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// Message is the upstream monitor payload.
type Message struct {
	ID          string    `json:"id"`
	Mailbox     string    `json:"mailbox"`
	From        string    `json:"from"`
	To          []string  `json:"to"`
	Subject     string    `json:"subject"`
	Date        time.Time `json:"date"`
	PosixMillis int64     `json:"posix-millis"`
	Size        int       `json:"size"`
	Seen        bool      `json:"seen"`
}

// SentAt returns the original send time, preferring Date and falling back to
// PosixMillis when Date is zero.
func (m Message) SentAt() time.Time {
	if !m.Date.IsZero() {
		return m.Date
	}
	if m.PosixMillis > 0 {
		return time.UnixMilli(m.PosixMillis)
	}
	return time.Time{}
}

// DisplayDate formats the original send time for the forwarded header. It
// returns an empty string when no timestamp is available.
func (m Message) DisplayDate() string {
	t := m.SentAt()
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC1123Z)
}

// Fingerprint returns a stable hash over the identifying header fields. The
// upstream monitor frame is a notification header (no body), so the fingerprint
// is computed over id+mailbox+from+to+subject. It is intentionally insensitive
// to Date / PosixMillis because WebSocket re-deliveries for the same logical
// email may arrive with different timestamps.
func (m Message) Fingerprint() string {
	h := sha256.New()
	h.Write([]byte(m.ID))
	h.Write([]byte("|"))
	h.Write([]byte(m.Mailbox))
	h.Write([]byte("|"))
	h.Write([]byte(m.From))
	h.Write([]byte("|"))
	h.Write([]byte(strings.Join(m.To, ",")))
	h.Write([]byte("|"))
	h.Write([]byte(m.Subject))
	return hex.EncodeToString(h.Sum(nil))
}
