// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"context"
	"fmt"

	"github.com/KeepShareOrg/keepshare/pkg/share"
)

// GetStatistics return the statistics of each host shared link.
func (p *PikPak) GetStatistics(ctx context.Context, userID string, hostSharedLinks []string) (details map[string]share.Statistics, err error) {
	details = make(map[string]share.Statistics, len(hostSharedLinks))
	for _, link := range hostSharedLinks {
		st, err := p.api.GetStatistics(ctx, link)
		if err != nil {
			return details, fmt.Errorf("get statistics err: %w", err)
		}
		details[link] = *st
	}
	return details, nil
}
