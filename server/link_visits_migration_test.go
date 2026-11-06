// Copyright 2026 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"os"
	"strings"
	"testing"

	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/rawsql"
)

func TestLinkVisitsExistingInstallMigrationMySQL(t *testing.T) {
	db := newLinkVisitsTestDB(t)
	for _, table := range []string{"keepshare_shared_link_visitor_ip", "keepshare_shared_link"} {
		if err := db.Exec("DROP TABLE " + table).Error; err != nil {
			t.Fatal(err)
		}
	}
	fresh, err := rawsql.FS.ReadFile("shared_link.sql")
	if err != nil {
		t.Fatal(err)
	}
	var oldSchema []string
	for _, line := range strings.Split(string(fresh), "\n") {
		if !strings.Contains(line, "`page_views`") && !strings.Contains(line, "`visitor_ips`") {
			oldSchema = append(oldSchema, line)
		}
	}
	if err := db.Exec(strings.Join(oldSchema, "\n")).Error; err != nil {
		t.Fatal(err)
	}
	seedLinkVisitTestRecord(t, db, 1)

	migration, err := os.ReadFile("../migrations/20261106_shared_link_visits.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range strings.Split(string(migration), ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	assertLinkVisitCounts(t, db, 1, 0, 0)
	var record model.SharedLink
	if err := db.First(&record, 1).Error; err != nil {
		t.Fatal(err)
	}
	if record.Visitor != 17 || record.Stored != 19 || record.Revenue != 23 || record.OriginalLink != "https://original.example/" {
		t.Error("migration changed historical record data")
	}
}
