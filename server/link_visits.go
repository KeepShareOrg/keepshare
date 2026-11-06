// Copyright 2026 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"net/netip"

	"gorm.io/gorm"
)

func recordLinkVisit(ctx context.Context, db *gorm.DB, linkID int64, ip string) error {
	if linkID <= 0 {
		return nil
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the parent first, serializing visits and deletion without changing updated_at.
		updated := tx.Exec("UPDATE keepshare_shared_link SET page_views = page_views + 1 WHERE auto_id = ?", linkID)
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 0 {
			return nil
		}

		hash, valid := linkVisitorIPHash(linkID, ip)
		if !valid {
			return nil
		}
		inserted := tx.Exec("INSERT IGNORE INTO keepshare_shared_link_visitor_ip (shared_link_id, ip_hash) VALUES (?, ?)", linkID, hash[:])
		if inserted.Error != nil {
			return inserted.Error
		}
		if inserted.RowsAffected == 0 {
			return nil
		}
		return tx.Exec("UPDATE keepshare_shared_link SET visitor_ips = visitor_ips + 1 WHERE auto_id = ?", linkID).Error
	})
}

func linkVisitorIPHash(linkID int64, ip string) ([sha256.Size]byte, bool) {
	addr, err := netip.ParseAddr(ip)
	if err != nil || addr.Zone() != "" {
		return [sha256.Size]byte{}, false
	}

	// Scope fingerprints to the record; equivalent IPv4 and mapped IPv6 share one key.
	var input [8 + 16]byte
	binary.BigEndian.PutUint64(input[:8], uint64(linkID))
	normalized := addr.Unmap().As16()
	copy(input[8:], normalized[:])
	return sha256.Sum256(input[:]), true
}
