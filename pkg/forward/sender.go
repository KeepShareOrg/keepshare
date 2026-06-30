// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package forward

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"mime/quotedprintable"
	"net/smtp"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
)

// Sender assembles and sends forwarded email. The notice is prepended to
// every forwarded body so the recipient knows the message is being relayed
// and was not authored by KeepShare.
type Sender struct {
	Notice string
}

// NewSender constructs a Sender with the given notice template. The notice
// template supports a single substitution token `<TO>`, replaced with the
// (un-masked) original PikPak master-account email at build time.
func NewSender(notice string) *Sender {
	return &Sender{Notice: notice}
}

// Built is the assembled MIME message ready for transport.
type Built struct {
	From    string
	To      string
	Subject string
	Body    []byte
}

// Build assembles the MIME payload. originalTo is the pikpak master account
// email the upstream message was addressed to (used to substitute `<TO>`).
// body is the plain-text body fetched from the mail server (the WebSocket
// notification frame carries only headers, never the body).
func (s *Sender) Build(account config.ForwardAccount, toUser, originalTo string, msg Message, body string) (*Built, error) {
	if account.Email == "" {
		return nil, errors.New("sender: account.email required")
	}
	if toUser == "" {
		return nil, errors.New("sender: toUser required")
	}
	boundary := fmt.Sprintf("%v", time.Now().UnixNano())
	// Substitute the forwarded-header tokens with the original message's
	// metadata so the recipient can see who sent it and when.
	notice := strings.NewReplacer(
		"<TO>", originalTo,
		"<FROM>", msg.From,
		"<DATE>", msg.DisplayDate(),
		"<SUBJECT>", msg.Subject,
	).Replace(s.Notice)

	text := notice
	if body != "" {
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += body
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s <%s>\r\n", account.Username, account.Email)
	fmt.Fprintf(&buf, "To: %s\r\n", toUser)
	fmt.Fprintf(&buf, "Subject: [Forward] %s\r\n", msg.Subject)
	buf.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: multipart/alternative; charset=UTF-8; boundary=%s\r\n", boundary)
	buf.WriteString("\r\n")

	// text/plain part
	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	buf.WriteString(base64.StdEncoding.EncodeToString([]byte(text)))
	buf.WriteString("\r\n")

	// text/html part (escaped notice + body, quoted-printable)
	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	var htmlBuf bytes.Buffer
	qp := quotedprintable.NewWriter(&htmlBuf)
	_, _ = qp.Write([]byte("<pre>" + html.EscapeString(text) + "</pre>"))
	_ = qp.Close()
	buf.Write(htmlBuf.Bytes())
	buf.WriteString("\r\n")

	buf.WriteString("--")
	buf.WriteString(boundary)
	buf.WriteString("--\r\n")

	return &Built{
		From:    account.Email,
		To:      toUser,
		Subject: "[Forward] " + msg.Subject,
		Body:    buf.Bytes(),
	}, nil
}

// Send delivers the message via the account's SMTP server with PlainAuth.
func (s *Sender) Send(b *Built, account config.ForwardAccount) error {
	if b == nil {
		return errors.New("sender: nil built message")
	}
	if account.Server == "" || account.Port == 0 {
		return errors.New("sender: account server/port required")
	}
	host := fmt.Sprintf("%s:%d", account.Server, account.Port)
	auth := smtp.PlainAuth("", account.Email, account.Password, account.Server)
	return smtp.SendMail(host, auth, b.From, []string{b.To}, b.Body)
}
