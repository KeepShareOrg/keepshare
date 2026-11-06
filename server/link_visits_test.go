// Copyright 2026 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/rawsql"
	drivermysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestLinkVisitorIPHash(t *testing.T) {
	for _, equivalent := range [][]string{
		{"192.0.2.1", "::ffff:192.0.2.1", "::ffff:c000:201"},
		{"2001:db8::1", "2001:0DB8:0000:0000:0000:0000:0000:0001"},
	} {
		want, valid := linkVisitorIPHash(42, equivalent[0])
		if !valid {
			t.Fatalf("valid IP rejected: %s", equivalent[0])
		}
		for _, ip := range equivalent {
			got, valid := linkVisitorIPHash(42, ip)
			if !valid || got != want {
				t.Errorf("equivalent IP %q produced a different fingerprint", ip)
			}
		}
		otherLink, _ := linkVisitorIPHash(43, equivalent[0])
		if otherLink == want {
			t.Error("fingerprints must be scoped to each sharing record")
		}
	}

	got, _ := linkVisitorIPHash(42, "192.0.2.1")
	otherIP, _ := linkVisitorIPHash(42, "192.0.2.2")
	if got == otherIP {
		t.Error("distinct IPs must have distinct fingerprints")
	}

	for _, ip := range []string{"", "invalid", "192.0.2.1:80", "[2001:db8::1]", "fe80::1%en0", "::ffff:192.0.2.1%eth0", "192.000.002.001", " 192.0.2.1", "192.0.2.1 ", "192.0.2.1, 192.0.2.2"} {
		if _, valid := linkVisitorIPHash(42, ip); valid {
			t.Errorf("invalid or zoned IP accepted: %q", ip)
		}
	}
}

func newLinkVisitsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("KEEPSHARE_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set KEEPSHARE_TEST_MYSQL_DSN to run isolated MySQL integration tests")
	}
	cfg, err := drivermysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	// Never use or mutate the supplied database, even if the DSN names one.
	cfg.DBName = ""
	cfg.ParseTime = true
	cfg.MultiStatements = false
	cfg.ClientFoundRows = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := admin.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	name := "keepshare_link_visits_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE `"+name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(ctx, "DROP DATABASE `"+name+"`"); err != nil {
			t.Errorf("clean temporary database %s: %v", name, err)
		}
	})

	cfg.DBName = name
	pool, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: pool, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	// Load the parent before its FK child, then the remaining fresh-install tables.
	files, err := rawsql.FS.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"shared_link.sql", "shared_link_visitor_ip.sql"}
	for _, file := range files {
		if file.Name() != "shared_link.sql" && file.Name() != "shared_link_visitor_ip.sql" {
			names = append(names, file.Name())
		}
	}
	for _, name := range names {
		stmt, err := rawsql.FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(stmt)).Error; err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
	}
	return db
}

func seedLinkVisitTestRecord(t *testing.T, db *gorm.DB, id int64) {
	t.Helper()
	err := db.Exec("INSERT INTO keepshare_shared_link "+
		"(auto_id, user_id, original_link_hash, host_shared_link_hash, original_link, host_shared_link, "+
		"visitor, `stored`, revenue, updated_at, first_visited_at, last_visited_at, last_stored_at) "+
		"VALUES (?, 'test', ?, ?, 'https://original.example/', 'https://host.example/', "+
		"17, 19, 23, '2020-01-02 03:04:05', '2020-02-03 04:05:06', '2020-03-04 05:06:07', '2020-04-05 06:07:08')",
		id, fmt.Sprintf("%040d", id), fmt.Sprintf("%040d", id)).Error
	if err != nil {
		t.Fatal(err)
	}
}

func assertLinkVisitCounts(t *testing.T, db *gorm.DB, id, pageViews, visitorIPs int64) {
	t.Helper()
	var record model.SharedLink
	if err := db.First(&record, id).Error; err != nil {
		t.Fatal(err)
	}
	if record.PageViews != pageViews || record.VisitorIps != visitorIPs {
		t.Errorf("record %d: got PV/IP %d/%d, want %d/%d", id, record.PageViews, record.VisitorIps, pageViews, visitorIPs)
	}
	var fingerprints int64
	if err := db.Table("keepshare_shared_link_visitor_ip").Where("shared_link_id = ?", id).Count(&fingerprints).Error; err != nil {
		t.Fatal(err)
	}
	if fingerprints != visitorIPs {
		t.Errorf("record %d: got %d fingerprints, want %d", id, fingerprints, visitorIPs)
	}
}

func TestRecordLinkVisitMySQL(t *testing.T) {
	db := newLinkVisitsTestDB(t)
	ctx := context.Background()
	seedLinkVisitTestRecord(t, db, 1)
	seedLinkVisitTestRecord(t, db, 2)
	assertLinkVisitCounts(t, db, 1, 0, 0)

	for _, ip := range []string{"192.0.2.1", "::ffff:192.0.2.1", "2001:db8::1", "2001:0DB8:0:0:0:0:0:1", "192.0.2.2", "", "invalid", "fe80::1%en0"} {
		if err := recordLinkVisit(ctx, db, 1, ip); err != nil {
			t.Fatal(err)
		}
	}
	assertLinkVisitCounts(t, db, 1, 8, 3)
	if err := recordLinkVisit(ctx, db, 2, "192.0.2.1"); err != nil {
		t.Fatal(err)
	}
	assertLinkVisitCounts(t, db, 2, 1, 1)

	var record model.SharedLink
	if err := db.First(&record, 1).Error; err != nil {
		t.Fatal(err)
	}
	if record.Visitor != 17 || record.Stored != 19 || record.Revenue != 23 ||
		record.UpdatedAt.Format(time.DateTime) != "2020-01-02 03:04:05" ||
		record.FirstVisitedAt.Format(time.DateTime) != "2020-02-03 04:05:06" ||
		record.LastVisitedAt.Format(time.DateTime) != "2020-03-04 05:06:07" ||
		record.LastStoredAt.Format(time.DateTime) != "2020-04-05 06:07:08" {
		t.Error("landing visits changed host statistics or original timestamps")
	}

	for _, id := range []int64{0, -1, 999} {
		if err := recordLinkVisit(ctx, db, id, "192.0.2.1"); err != nil {
			t.Fatal(err)
		}
	}
	assertLinkVisitCounts(t, db, 1, 8, 3)
	assertLinkVisitCounts(t, db, 2, 1, 1)
	var records int64
	if err := db.Model(&model.SharedLink{}).Count(&records).Error; err != nil {
		t.Fatal(err)
	}
	if records != 2 {
		t.Error("invalid or missing IDs changed the sharing records")
	}
	var orphanIPs int64
	if err := db.Table("keepshare_shared_link_visitor_ip").Where("shared_link_id NOT IN (1, 2)").Count(&orphanIPs).Error; err != nil {
		t.Fatal(err)
	}
	if orphanIPs != 0 {
		t.Error("missing records created orphan fingerprints")
	}
	if err := db.Delete(&model.SharedLink{}, 2).Error; err != nil {
		t.Fatal(err)
	}
	if err := recordLinkVisit(ctx, db, 2, "192.0.2.1"); err != nil {
		t.Fatal(err)
	}
	var deletedIPs int64
	if err := db.Table("keepshare_shared_link_visitor_ip").Where("shared_link_id = 2").Count(&deletedIPs).Error; err != nil {
		t.Fatal(err)
	}
	if deletedIPs != 0 {
		t.Error("hard deletion left fingerprints behind")
	}
}

func TestRecordLinkVisitConcurrentMySQL(t *testing.T) {
	db := newLinkVisitsTestDB(t)
	seedLinkVisitTestRecord(t, db, 1)
	const visits = 48
	start := make(chan struct{})
	errors := make(chan error, visits)
	var workers sync.WaitGroup
	for i := range visits {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			<-start
			errors <- recordLinkVisit(context.Background(), db, 1, fmt.Sprintf("192.0.2.%d", i%4+1))
		}(i)
	}
	close(start)
	workers.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
	assertLinkVisitCounts(t, db, 1, visits, 4)
}

func TestRecordLinkVisitRollbackMySQL(t *testing.T) {
	db := newLinkVisitsTestDB(t)
	seedLinkVisitTestRecord(t, db, 1)
	if err := db.Exec("DROP TABLE keepshare_shared_link_visitor_ip").Error; err != nil {
		t.Fatal(err)
	}
	if err := recordLinkVisit(context.Background(), db, 1, "192.0.2.1"); err == nil {
		t.Fatal("expected deduplication failure")
	}
	var record model.SharedLink
	if err := db.First(&record, 1).Error; err != nil {
		t.Fatal(err)
	}
	if record.PageViews != 0 || record.VisitorIps != 0 {
		t.Error("deduplication failure did not roll back both counters")
	}
}
