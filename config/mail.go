// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"

	"github.com/KeepShareOrg/keepshare/pkg/mail"
	"github.com/KeepShareOrg/keepshare/pkg/mail/inbucket"
	log "github.com/sirupsen/logrus"
)

var mailer mail.Mailer

func initMail() error {
	cli, err := inbucket.New(mailServer())
	if err != nil {
		return fmt.Errorf("init mail client err: %w", err)
	}
	mailer = cli
	return nil
}

// Mailer returns the mail client.
func Mailer() mail.Mailer {
	if mailer == nil {
		log.Fatal("mail client has not been initialized")
	}
	return mailer
}
