// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package forward

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Deduper prevents replaying the same logical email across WebSocket
// re-deliveries or keepshare instance re-sends. Fingerprints live in Redis
// for ttl (default 24h), so dedup state is shared across replicas.
type Deduper struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewDeduper returns a Deduper writing under the forward:dedup:* namespace.
func NewDeduper(rdb *redis.Client, ttl time.Duration) *Deduper {
	return &Deduper{rdb: rdb, ttl: ttl}
}

// FirstSeen returns true iff this fingerprint was not present within TTL.
func (d *Deduper) FirstSeen(ctx context.Context, fp string) (bool, error) {
	ok, err := d.rdb.SetNX(ctx, "forward:dedup:"+fp, 1, d.ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// Forget removes the fingerprint, allowing a future delivery to be processed.
func (d *Deduper) Forget(ctx context.Context, fp string) error {
	return d.rdb.Del(ctx, "forward:dedup:"+fp).Err()
}
