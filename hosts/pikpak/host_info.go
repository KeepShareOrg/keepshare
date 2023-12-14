// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"context"

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
