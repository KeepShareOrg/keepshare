package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	pikpakmodel "github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/magnetblock"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/model"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newMagnetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("KS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set KS_TEST_MYSQL_DSN to run isolated MySQL regressions (CREATE DATABASE required)")
	}
	options, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid KS_TEST_MYSQL_DSN")
	}
	options.DBName = ""
	options.ParseTime = true
	admin, err := gorm.Open(mysql.Open(options.FormatDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot open test MySQL")
	}
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { adminSQL.Close() })
	name := "ks_magnet_test_" + uuid.NewString()[:8]
	if err := admin.Exec("CREATE DATABASE `" + name + "`").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec("DROP DATABASE `" + name + "`").Error; err != nil {
			t.Error(err)
		}
	})
	options.DBName = name
	db, err := gorm.Open(mysql.Open(options.FormatDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	for _, path := range []string{
		"rawsql/shared_link.sql", "rawsql/blacklist.sql", "rawsql/magnet_blacklist.sql",
		"../hosts/pikpak/rawsql/shared_link.sql", "../hosts/pikpak/rawsql/file.sql",
	} {
		statement, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(statement)).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestMagnetTaskKeyMigrationPreservesExistingKeys(t *testing.T) {
	db := newMagnetTestDB(t)
	if err := db.Exec("ALTER TABLE pikpak_file MODIFY unique_hash char(50) NOT NULL").Error; err != nil {
		t.Fatal(err)
	}
	const hash = "0123456789abcdef0123456789abcdef01234567"
	const user = "1234567890abcdef"
	old := &pikpakmodel.File{
		MasterUserID: "old-master", WorkerUserID: "old-worker", TaskID: "old-task",
		OriginalLinkHash: magnetblock.Aliases(hash)[1], UniqueHash: user + ":" + magnetblock.Aliases(hash)[1],
	}
	if err := db.Create(old).Error; err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../migrations/20261106_pikpak_file_unique_hash.sql")
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := db.Exec(string(migration)).Error; err != nil {
			t.Fatal(err)
		}
	}
	current := &pikpakmodel.File{
		MasterUserID: "new-master", WorkerUserID: "new-worker", TaskID: "new-task",
		OriginalLinkHash: hash, UniqueHash: user + ":" + hash,
	}
	if err := db.Create(current).Error; err != nil {
		t.Fatalf("canonical key for a maximum-length user must fit: %v", err)
	}
	var storedOld, storedNew pikpakmodel.File
	if err := db.First(&storedOld, old.AutoID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&storedNew, current.AutoID).Error; err != nil {
		t.Fatal(err)
	}
	if storedOld.UniqueHash != old.UniqueHash || storedNew.UniqueHash != current.UniqueHash {
		t.Fatal("migration changed an existing key or truncated the canonical key")
	}
	duplicate := *current
	duplicate.AutoID = 0
	duplicate.MasterUserID, duplicate.WorkerUserID = "other-master", "other-worker"
	if err := db.Create(&duplicate).Error; err == nil {
		t.Fatal("key migration removed uniqueness protection")
	}
}

func TestRevokedNativeShareCannotBePublishedAfterUnblock(t *testing.T) {
	db := newMagnetTestDB(t)
	ctx := context.Background()
	const hash = "0123456789abcdef0123456789abcdef01234567"
	const link = "https://mypikpak.invalid/s/old-share?act=play"
	record := &model.SharedLink{UserID: "user", Host: "pikpak", OriginalLinkHash: hash,
		OriginalLink: "magnet:?xt=urn:btih:" + hash, State: share.StatusCreated.String()}
	if err := db.Create(record).Error; err != nil {
		t.Fatal(err)
	}
	native := &pikpakmodel.SharedLink{ShareID: "old-share", WorkerUserID: "worker", FileID: "file"}
	if err := db.Create(native).Error; err != nil {
		t.Fatal(err)
	}
	block := &model.MagnetBlacklist{InfoHash: hash, OriginalLink: record.OriginalLink}
	if err := db.Create(block).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(native).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(block).Error; err != nil {
		t.Fatal(err)
	}
	publish := &model.SharedLink{State: share.StatusOK.String(), HostSharedLink: link}
	if err := updateSharedLink(db.WithContext(ctx), record.UserID, record.Host, hash, publish); err != nil {
		t.Fatal(err)
	}
	var got model.SharedLink
	if err := db.First(&got, record.AutoID).Error; err != nil {
		t.Fatal(err)
	}
	if got.State != share.StatusCreated.String() || got.HostSharedLink != "" {
		t.Fatalf("revoked callback resurrected the record: %s %s", got.State, got.HostSharedLink)
	}
	// A genuinely new native share can still complete after the block is removed.
	if err := db.Create(&pikpakmodel.SharedLink{ShareID: "new-share", WorkerUserID: "worker", FileID: "file"}).Error; err != nil {
		t.Fatal(err)
	}
	publish.HostSharedLink = "https://mypikpak.invalid/s/new-share"
	if err := updateSharedLink(db.WithContext(ctx), record.UserID, record.Host, hash, publish); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&got, record.AutoID).Error; err != nil {
		t.Fatal(err)
	}
	if got.State != share.StatusOK.String() || got.HostSharedLink != publish.HostSharedLink {
		t.Fatal("a valid post-unblock share could not complete")
	}
}

func TestNativePublicationRejectsConcurrentlyDeletedMapping(t *testing.T) {
	db := newMagnetTestDB(t)
	const hash = "0123456789abcdef0123456789abcdef01234567"
	record := &model.SharedLink{UserID: "user", Host: "pikpak", OriginalLinkHash: hash, State: share.StatusCreated.String()}
	native := &pikpakmodel.SharedLink{ShareID: "revoking-share", WorkerUserID: "worker", FileID: "file"}
	if err := db.Create(record).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(native).Error; err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	if err := tx.Delete(native).Error; err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		finished <- updateSharedLink(db.WithContext(ctx), record.UserID, record.Host, hash,
			&model.SharedLink{State: share.StatusOK.String(), HostSharedLink: "https://mypikpak.invalid/s/revoking-share"})
	}()
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	var got model.SharedLink
	if err := db.First(&got, record.AutoID).Error; err != nil {
		t.Fatal(err)
	}
	if got.State != share.StatusCreated.String() || got.HostSharedLink != "" {
		t.Fatal("publication used a native mapping deleted by a concurrent revocation")
	}
}
