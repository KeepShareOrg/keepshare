// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package mail

import (
	"context"
	"time"
)

// Mailer is a mail client to send and get mails.
type Mailer interface {
	// List returns a list of message headers for the requested address.
	List(ctx context.Context, address string) ([]*Header, error)

	// Get returns the message body by a address name and message ID.
	Get(ctx context.Context, address string, id string) (*Body, error)

	// Del deletes a single message given the address name and message ID.
	Del(ctx context.Context, address string, id string) error

	// Clear deletes all messages in the given address.
	Clear(ctx context.Context, address string) error

	// Domain returns the domain of the mailer.
	Domain() string
}

// Header is the header of a mail.
type Header struct {
	Address string    `json:"address"`
	ID      string    `json:"id"`
	From    string    `json:"from"`
	To      []string  `json:"to"`
	Subject string    `json:"subject"`
	Date    time.Time `json:"date"`
	Size    int64     `json:"size"`
	Seen    bool      `json:"seen"`
}

// Body is the body of a mail.
type Body struct {
	Text string `json:"text"`
	HTML string `json:"html"`
}
