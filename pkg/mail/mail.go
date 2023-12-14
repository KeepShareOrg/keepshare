// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package mail

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/KeepShareOrg/keepshare/pkg/log"
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

// Filter options to filter emails.
type Filter struct {
	SendTime      time.Time      // if not zero, filter mails received after this time.
	From          string         // if not empty, filter emails FROM equal to this value.
	FromRegexp    *regexp.Regexp // if not nil, filter emails FROM matching this regular expression.
	Subject       string         // is not empty, filter emails SUBJECT equal to this value.
	SubjectRegexp *regexp.Regexp // if not nil, filter emails SUBJECT matching this regular expression.
}

// FindText find text from emails.
// If textRegexp is nil, return the body of the latest email which matches the headerFilter.
func FindText(ctx context.Context, mailer Mailer, email string, textRegexp *regexp.Regexp, f *Filter) (text string, found bool, err error) {
	if mailer == nil {
		return "", false, errors.New("nil mailer")
	}
	if email == "" {
		return "", false, errors.New("email is required")
	}

	headers, err := mailer.List(ctx, email)
	if err != nil {
		return "", false, fmt.Errorf("list mail err: %w", err)
	}

	log.WithContext(ctx).Debugf("list headers for mail: %s, length: %d", email, len(headers))

	// The newest emails are at the end of the list.
	for i := len(headers) - 1; i >= 0; i-- {
		h := headers[i]
		log.WithContext(ctx).Debugf("mail header[%d]: %+v", i, h)

		if f.SendTime.Year() >= 2023 && h.Date.Before(f.SendTime) {
			continue
		}

		if f.From != "" && f.From != h.From {
			continue
		}

		if f.FromRegexp != nil && !f.FromRegexp.MatchString(h.From) {
			continue
		}

		if f.Subject != "" && f.Subject != h.Subject {
			continue
		}

		if f.SubjectRegexp != nil && !f.SubjectRegexp.MatchString(h.Subject) {
			continue
		}

		body, err := mailer.Get(ctx, email, h.ID)
		if err != nil {
			return "", false, fmt.Errorf("get mail body err: %w", err)
		}

		log.WithContext(ctx).Debugf("mail text body[%d]: %s", i, body.Text)

		if textRegexp == nil {
			return body.Text, true, nil
		}

		text = textRegexp.FindString(body.Text)
		log.WithContext(ctx).Debugf("text regexp: `%s`, match result: `%s`", textRegexp.String(), text)
		if text != "" {
			// clear mail records after find code.
			mailer.Clear(ctx, email)
			return text, true, nil
		}
	}

	return "", false, nil
}
