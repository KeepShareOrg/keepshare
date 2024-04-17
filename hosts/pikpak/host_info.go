// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"context"
	"fmt"
	"github.com/KeepShareOrg/keepshare/pkg/log"
)

// HostInfo returns basic information of the host.
func (p *PikPak) HostInfo(ctx context.Context, userID string, options map[string]any) (resp map[string]any, err error) {
	resp = make(map[string]any)

	// get master account.
	master, err := p.m.GetMaster(ctx, userID)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("get master err")
		return nil, err
	}
	resp["master"] = master

	// get worker accounts, free and premium, storage limit and used
	workers, err := p.m.CountWorkers(ctx, master.UserID)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("count workers err")
		return nil, err
	}
	resp["workers"] = workers

	// get total revenue, no result is returned if an error occurs.
	commission, err := p.api.GetCommissions(ctx, master.UserID)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("get revenue err")
	} else {
		resp["revenue"] = commission.Total
	}

	return resp, nil
}

// ChangeMasterAccountPassword changes the master account password.
func (p *PikPak) ChangeMasterAccountPassword(ctx context.Context, userID, newPassword string, rememberMe bool) (string, error) {
	// query the master account registration email
	registrationEmail, err := p.q.MasterAccount.WithContext(ctx).
		Select(p.q.MasterAccount.Email).
		Where(p.q.MasterAccount.KeepshareUserID.Eq(userID)).
		Take()
	if err != nil {
		return "", fmt.Errorf("query master account email error, %v", err.Error())
	}
	log.Debugf("master account email is %s", registrationEmail.Email)

	taskID, err := p.api.ResetPassword(ctx, registrationEmail.Email, newPassword, rememberMe)
	if err != nil {
		return "", err
	}
	return taskID, nil
}

func (p *PikPak) ConfirmMasterAccountPassword(ctx context.Context, keepShareUserID, password string, savePassword bool) error {
	return p.api.ConfirmPassword(ctx, keepShareUserID, password, savePassword)
}

const (
	LoginStatusValid   = "valid"
	LoginStatusInvalid = "invalid"
)

func (p *PikPak) GetMasterAccountLoginStatus(ctx context.Context, keepShareUserID string) (string, error) {
	tk := p.q.Token
	ma := p.q.MasterAccount

	loginStatus := LoginStatusInvalid
	var res struct {
		Password    string
		AccessToken string
	}
	err := ma.WithContext(ctx).Select(ma.Password, tk.AccessToken).LeftJoin(tk, ma.UserID.EqCol(tk.UserID)).Where(ma.KeepshareUserID.Eq(keepShareUserID)).Scan(&res)
	if err != nil {
		return loginStatus, err
	}

	if res.Password != "" || res.AccessToken != "" {
		loginStatus = LoginStatusValid
	}

	return loginStatus, nil
}
