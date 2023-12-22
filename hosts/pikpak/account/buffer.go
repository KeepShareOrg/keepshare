// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package account

import (
	"context"
	"fmt"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/KeepShareOrg/keepshare/pkg/util"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/spf13/viper"
)

const rateLimitedError = "|aborted|"

func (m *Manager) initConfig() {
	if v := viper.GetInt("pikpak.master_buffer_size"); v > 0 {
		m.masterBufferSize = v
	}
	if v := viper.GetInt("pikpak.master_buffer_concurrency"); v > 0 {
		m.masterBufferConcurrency = v
	}
	if v := viper.GetDuration("pikpak.master_buffer_interval"); v > 0 {
		m.masterBufferInterval = v
	}

	if v := viper.GetInt("pikpak.worker_buffer_size"); v > 0 {
		m.workerBufferSize = v
	}
	if v := viper.GetInt("pikpak.worker_buffer_concurrency"); v > 0 {
		m.workerBufferConcurrency = v
	}
	if v := viper.GetDuration("pikpak.worker_buffer_interval"); v > 0 {
		m.workerBufferInterval = v
	}
}

func (m *Manager) checkMasterBuffer() {
	const timeout = 30 * time.Second

	do := func() error {
		t := &m.q.MasterAccount
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		count, err := t.WithContext(gormutil.IgnoreTraceContext(ctx)).Where(t.KeepshareUserID.Eq("")).Count()
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("count master buffer err")
			return err
		}

		if count >= int64(m.masterBufferSize) {
			time.Sleep(m.masterBufferInterval)
			return nil
		}

		var wg sync.WaitGroup
		var we error
		for i := 0; i < m.masterBufferConcurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// register get a new account with random email.
				user, err := m.api.SignUp(ctx, "", timeout)
				if err != nil {
					log.WithContext(ctx).WithError(err).Error("sign up err")
					we = err
					return
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
					log.WithContext(ctx).WithError(err).Error("create master account err")
					we = err
					return
				}
			}()
		}

		wg.Wait()
		return we
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
			log.WithContext(ctx).WithError(err).Error("count worker buffer err")
			return err
		}

		if count >= int64(m.workerBufferSize) {
			time.Sleep(m.workerBufferInterval)
			return nil
		}

		var wg sync.WaitGroup
		var we error
		for i := 0; i < m.workerBufferConcurrency; i++ {
			wg.Add(1)
			if i > 0 {
				time.Sleep(time.Second)
			}
			go func() {
				defer wg.Done()

				// register get a new account with random email.
				user, err := m.api.SignUp(ctx, "", timeout)
				if err != nil {
					log.WithContext(ctx).WithError(err).Error("sign up err")
					we = err
					return
				}

				now := time.Now()
				a := &model.WorkerAccount{
					UserID:            user.UserID,
					MasterUserID:      "",
					Email:             user.Email,
					Password:          user.Password,
					UsedSize:          0,
					LimitSize:         6 * util.GB,
					PremiumExpiration: time.Time{},
					CreatedAt:         now,
					UpdatedAt:         now,
				}

				if err = t.WithContext(ctx).Create(a); err != nil {
					log.WithContext(ctx).WithError(err).Error("create worker account err")
					we = err
					return
				}
			}()
		}

		wg.Wait()
		return we
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

// checkPremiumWorkerBuffer check the not used premium workers and redeem if they less than the limit
func (m *Manager) checkPremiumWorkerBuffer() {
	ctx := context.Background()
	l := log.WithContext(ctx)

	t := m.q
	do := func() error {
		// get the not used premium workers count
		notUsedPremiumCount, err := t.WorkerAccount.WithContext(ctx).
			Where(
				t.WorkerAccount.MasterUserID.Eq(""),
				t.WorkerAccount.PremiumExpiration.Lt(time.Now()), // count free or expired workers only
			).Count()
		if err != nil {
			return fmt.Errorf("count premium worker buffer err: %v", err)
		}

		if notUsedPremiumCount < int64(m.premiumBufferSize) {
			// select a not used normal worker account
			notUsedNormalAccount, err := t.WorkerAccount.WithContext(ctx).
				Where(
					t.WorkerAccount.MasterUserID.Eq(""),
					t.WorkerAccount.PremiumExpiration.Lt(time.Now()),
				).Take()
			if err != nil {
				return fmt.Errorf("found not use normal account err: %v", err)
			}

			// select a not used redeem code
			notUsedRedeemCode, err := t.RedeemCode.WithContext(ctx).
				Where(t.RedeemCode.Status.Eq(comm.RedeemCodeStatusNotUsed)).
				Order().
				Take()

			l.Debugf("not used normal account, code: %v, %v", notUsedNormalAccount, notUsedRedeemCode)
			if err != nil {
				return fmt.Errorf("found not use redeem code err: %v", err)
			}

			// redeem
			err = m.api.Redeem(ctx, notUsedNormalAccount.UserID, notUsedRedeemCode.Code)
			if err != nil {
				if api.IsInvalidRedeemCodeErr(err) {
					// mark the redeem code as invalid
					t.RedeemCode.WithContext(ctx).
						Where(t.RedeemCode.AutoID.Eq(notUsedRedeemCode.AutoID)).
						Update(t.RedeemCode.Status, comm.RedeemCodeInvalid)

					return nil
				}
				return fmt.Errorf("redeem err: %v", err)
			}

			// mark the redeem code as used
			t.RedeemCode.WithContext(ctx).
				Where(t.RedeemCode.AutoID.Eq(notUsedRedeemCode.AutoID)).
				Updates(&model.RedeemCode{
					UsedUserID: notUsedNormalAccount.UserID,
					Status:     comm.RedeemCodeStatusUsed,
				})

			// update the account premium expiration info
			m.api.UpdateWorkerPremium(ctx, notUsedNormalAccount)
		}

		return nil
	}

	for {
		if err := do(); err != nil {
			if strings.Contains(err.Error(), rateLimitedError) {
				time.Sleep(time.Duration(60+rand.Int31n(30)) * time.Second)
			}
		}
		time.Sleep(m.premiumBufferInterval)
	}
}
