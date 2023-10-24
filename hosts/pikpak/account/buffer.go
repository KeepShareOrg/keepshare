// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package account

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const rateLimitedError = "|aborted|"

func (m *Manager) initConfig() {
	if v := viper.GetInt("pikpak.master_buffer_size"); v > 0 {
		m.masterBufferSize = v
	}
	if v := viper.GetDuration("pikpak.master_buffer_interval"); v > 0 {
		m.masterBufferInterval = v
	}
	if v := viper.GetInt("pikpak.worker_buffer_size"); v > 0 {
		m.workerBufferSize = v
	}
	if v := viper.GetDuration("pikpak.worker_buffer_interval"); v > 0 {
		m.workerBufferInterval = v
	}
}

func (m *Manager) checkMasterBuffer() {
	const timeout = 20 * time.Second

	do := func() error {
		t := &m.q.MasterAccount
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		count, err := t.WithContext(gormutil.IgnoreTraceContext(ctx)).Where(t.KeepshareUserID.Eq("")).Count()
		if err != nil {
			log.WithError(err).Error("count master buffer err")
			return err
		}

		if count >= int64(m.masterBufferSize) {
			time.Sleep(m.masterBufferInterval)
			return nil
		}

		// register get a new account with random email.
		user, err := m.api.SignUp(ctx, "", timeout)
		if err != nil {
			log.WithError(err).Error("sign up err")
			return err
		}

		now := time.Now()
		a := &model.MasterAccount{
			UserID:          user.UserID,
			KeepshareUserID: "",
			Email:           user.Email,
			Password:        user.Password,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		if err = t.WithContext(ctx).Create(a); err != nil {
			log.WithError(err).Error("create master account err")
			return err
		}

		return nil
	}

	for {
		if err := do(); err != nil {
			time.Sleep(5 * m.masterBufferInterval)
			if strings.Contains(err.Error(), rateLimitedError) {
				time.Sleep(time.Duration(60+rand.Int31n(30)) * time.Second)
			}
		}
	}
}

func (m *Manager) checkWorkerBuffer() {
	const timeout = 60 * time.Second

	do := func() error {
		t := &m.q.WorkerAccount
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		count, err := t.WithContext(gormutil.IgnoreTraceContext(ctx)).Where(
			t.MasterUserID.Eq(""),
			t.PremiumExpiration.Lt(time.Now()), // count free or expired workers only
		).Count()
		if err != nil {
			log.WithError(err).Error("count worker buffer err")
			return err
		}

		if count >= int64(m.workerBufferSize) {
			time.Sleep(m.workerBufferInterval)
			return nil
		}

		// register get a new account with random email.
		user, err := m.api.SignUp(ctx, "", timeout)
		if err != nil {
			log.WithError(err).Error("sign up err")
			return err
		}

		now := time.Now()
		a := &model.WorkerAccount{
			UserID:            user.UserID,
			MasterUserID:      "",
			Email:             user.Email,
			Password:          user.Password,
			UsedSize:          0,
			LimitSize:         0,
			PremiumExpiration: time.Time{},
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err = t.WithContext(ctx).Create(a); err != nil {
			log.WithError(err).Error("create worker account err")
			return err
		}

		return nil
	}

	for {
		if err := do(); err != nil {
			time.Sleep(5 * m.workerBufferInterval)
			if strings.Contains(err.Error(), rateLimitedError) {
				time.Sleep(time.Duration(60+rand.Int31n(30)) * time.Second)
			}
		}
	}
}
