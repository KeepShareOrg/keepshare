// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package inbucket

import (
	"context"
	"fmt"
	"net/url"

	"github.com/KeepShareOrg/keepshare/pkg/mail"
	"github.com/inbucket/inbucket/pkg/rest/client"
)

// Client of an inbucket server.
type Client struct {
	client *client.Client
	domain string
}

// New creates a new client of an inbucket server.
func New(baseURL string) (mail.Mailer, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse mailer base url err: %w", err)
	}

	domain := u.Hostname()
	if domain == "" {
		return nil, fmt.Errorf("parse mailer base url err: invalid domain")
	}

	cli, err := client.New(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{client: cli, domain: domain}, nil
}

// List returns a list of message headers for the requested address.
func (c *Client) List(_ context.Context, address string) ([]*mail.Header, error) {
	hs, err := c.client.ListMailbox(address)
	if err != nil {
		return nil, err
	}

	hs2 := make([]*mail.Header, 0, len(hs))
	for _, v := range hs {
		hs2 = append(hs2, &mail.Header{
			Address: v.Mailbox,
			ID:      v.ID,
			From:    v.From,
			To:      v.To,
			Subject: v.Subject,
			Date:    v.Date,
			Size:    v.Size,
			Seen:    v.Seen,
		})
	}
	return hs2, nil
}

// Get returns the message body by a address name and message ID.
func (c *Client) Get(_ context.Context, address string, id string) (*mail.Body, error) {
	msg, err := c.client.GetMessage(address, id)
	if err != nil {
		return nil, err
	}
	return &mail.Body{
		Text: msg.Body.Text,
		HTML: msg.Body.HTML,
	}, nil
}

// Del deletes a single message given the address name and message ID.
func (c *Client) Del(ctx context.Context, address string, id string) error {
	return c.client.DeleteMessage(address, id)
}

// Clear deletes all messages in the given address.
func (c *Client) Clear(_ context.Context, address string) error {
	return c.client.PurgeMailbox(address)
}

// Domain returns the domain of the mailer.
func (c *Client) Domain() string {
	return c.domain
}
