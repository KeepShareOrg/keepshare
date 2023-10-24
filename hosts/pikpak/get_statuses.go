// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"context"

	"github.com/KeepShareOrg/keepshare/pkg/share"
)

// GetStatuses return the statuses of each host shared link.
func (p *PikPak) GetStatuses(ctx context.Context, _ string, hostSharedLinks []string) (statuses map[string]share.State, err error) {
	statuses = make(map[string]share.State, len(hostSharedLinks))
	for _, link := range hostSharedLinks {
		status, _, err := p.api.GetShareStatus(ctx, link)
		if err != nil {
			return statuses, err
		}
		statuses[link] = status
	}
	return statuses, nil
}
