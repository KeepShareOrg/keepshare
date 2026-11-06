package api

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/magnetblock"
	"github.com/coocood/freecache"
	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const testMagnetHash = "0123456789abcdef0123456789abcdef01234567"

func TestFileTaskCanonicalizesMagnetHashOnly(t *testing.T) {
	task := &fileTask{ID: "task", FileID: "file"}
	for _, input := range []string{
		"magnet:?xt=urn:sha1:other&xt=urn:btih:" + strings.ToUpper(testMagnetHash),
		"magnet:?xt=urn:btih:" + magnetblock.Aliases(testMagnetHash)[1],
	} {
		file := task.toFile("user", "master", "worker", input)
		if file.OriginalLinkHash != testMagnetHash || file.UniqueHash != "user:"+testMagnetHash {
			t.Fatalf("magnet source hash = %q, unique hash = %q", file.OriginalLinkHash, file.UniqueHash)
		}
	}
	for _, input := range []string{"https://example.invalid/file", "ftp://example.invalid/file"} {
		if got := task.toFile("user", "master", "worker", input).OriginalLinkHash; got != lk.Hash(input) {
			t.Fatalf("non-magnet source hash changed to %q", got)
		}
	}
}

func TestCreateFilesFromLinkRejectsBlockedBeforeNetwork(t *testing.T) {
	db := newScriptDB(t, sqlStep{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}, rows: [][]driver.Value{{testMagnetHash}}})
	api := testAPI(db, "worker")
	stubPikPak(t, func(r *http.Request) (int, string) {
		t.Fatalf("blocked source issued %s", r.URL.Path)
		return 500, ""
	})
	file, err := api.CreateFilesFromLink(context.Background(), "master", "worker", "magnet:?xt=urn:btih:"+testMagnetHash)
	if file != nil || !errors.Is(err, magnetblock.ErrBlocked) {
		t.Fatalf("CreateFilesFromLink() = %v, %v", file, err)
	}
}

func TestCreateFilesFromLinkRechecksAndKeepsCloudFiles(t *testing.T) {
	db := newScriptDB(t,
		sqlStep{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}},
		sqlStep{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}},
		sqlStep{query: "pikpak_master_account", columns: []string{"keepshare_user_id"}, rows: [][]driver.Value{{"user"}}},
		sqlStep{exec: "INSERT INTO `pikpak_file`"},
		sqlStep{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}, rows: [][]driver.Value{{testMagnetHash}}},
	)
	api := testAPI(db, "worker")
	var requests int
	stubPikPak(t, func(r *http.Request) (int, string) {
		requests++
		if r.URL.Path != "/drive/v1/files" {
			t.Fatalf("newly blocked files were deleted: %s", r.URL.Path)
		}
		return 200, `{"task":{"id":"task","file_id":"file","phase":"PHASE_TYPE_COMPLETE","file_name":"file","file_size":"1"}}`
	})
	file, err := api.CreateFilesFromLink(context.Background(), "master", "worker", "magnet:?xt=urn:btih:"+testMagnetHash)
	if file != nil || !errors.Is(err, magnetblock.ErrBlocked) || requests != 1 {
		t.Fatalf("CreateFilesFromLink() = %v, %v; cloud operations = %d", file, err, requests)
	}
}

func TestCreateShareRejectsDatabaseFailureBeforeNetwork(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	db := newScriptDB(t, sqlStep{query: "pikpak_file", err: databaseErr})
	api := testAPI(db, "worker")
	stubPikPak(t, func(r *http.Request) (int, string) {
		t.Fatalf("failed source lookup issued %s", r.URL.Path)
		return 500, ""
	})
	link, err := api.CreateShare(context.Background(), "master", "worker", "file")
	if link != "" || !errors.Is(err, databaseErr) {
		t.Fatalf("CreateShare() = %q, %v", link, err)
	}
}

func TestCreateShareRejectsBlockedBeforeCredentialLookup(t *testing.T) {
	db := newScriptDB(t,
		sqlStep{query: "pikpak_file", columns: []string{"original_link_hash"}, rows: [][]driver.Value{{testMagnetHash}}},
		sqlStep{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}, rows: [][]driver.Value{{testMagnetHash}}},
	)
	api := testAPI(db)
	stubPikPak(t, func(r *http.Request) (int, string) {
		t.Fatalf("blocked source issued %s", r.URL.Path)
		return 500, ""
	})
	link, err := api.CreateShare(context.Background(), "master", "worker", "file")
	if link != "" || !errors.Is(err, magnetblock.ErrBlocked) {
		t.Fatalf("CreateShare() = %q, %v", link, err)
	}
}

func TestCreateShareRechecksAfterCreationAndPreservesFailedRevocation(t *testing.T) {
	for _, revokeStatus := range []int{204, 503} {
		t.Run(http.StatusText(revokeStatus), func(t *testing.T) {
			steps := []sqlStep{
				{query: "pikpak_file", columns: []string{"original_link_hash"}, rows: [][]driver.Value{{testMagnetHash}}},
				{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}},
				{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}},
				{exec: "INSERT INTO `pikpak_shared_link`"},
				{query: "keepshare_magnet_blacklist", columns: []string{"info_hash"}, rows: [][]driver.Value{{testMagnetHash}}},
			}
			if revokeStatus == 204 {
				steps = append(steps, sqlStep{exec: "DELETE FROM `pikpak_shared_link`"})
			}
			db := newScriptDB(t, steps...)
			api := testAPI(db, "worker")
			var requests []string
			stubPikPak(t, func(r *http.Request) (int, string) {
				requests = append(requests, r.URL.Path)
				if r.URL.Path == "/drive/v1/share" {
					return 200, `{"share_url":"https://mypikpak.com/s/new-share"}`
				}
				if r.URL.Path == "/drive/v1/share:batchDelete" {
					return revokeStatus, ""
				}
				t.Fatalf("unexpected cloud operation %s", r.URL.Path)
				return 500, ""
			})
			link, err := api.CreateShare(context.Background(), "master", "worker", "file")
			if link != "" || !errors.Is(err, magnetblock.ErrBlocked) {
				t.Fatalf("CreateShare() published %q, %v", link, err)
			}
			if len(requests) != 2 {
				t.Fatalf("cloud operations = %v; want create and share-only revocation", requests)
			}
		})
	}
}

func TestDeleteShareRejectsHTTPAndProviderErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"empty HTTP error", 503, ""},
		{"HTML HTTP error", 403, "<html>denied</html>"},
		{"JSON HTTP error", 401, `{"error":"unauthorized","error_description":"fake-token","error_code":16}`},
		{"success status with error", 200, `{"error":"rejected","error_description":"fake-token","error_code":16}`},
		{"success status with error code only", 200, `{"error_code":16}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			api := testAPI(nil, "worker")
			stubPikPak(t, func(*http.Request) (int, string) { return test.status, test.body })
			err := api.DeleteShare(context.Background(), "worker", []string{"share"})
			if err == nil || strings.Contains(err.Error(), "fake-token") {
				t.Fatalf("DeleteShare() returned an unsafe success/error: %v", err)
			}
		})
	}
}

func TestRevokeMagnetSharesFindsOldURLsWithoutFilesAndContinuesOtherWorkers(t *testing.T) {
	db := newScriptDB(t,
		sqlStep{query: "JOIN pikpak_file", columns: []string{"share_id"}},
		sqlStep{query: "keepshare_shared_link", columns: []string{"host_shared_link"}, rows: [][]driver.Value{
			{"https://mypikpak.com/s/old-share?act=play"}, {"https://mypikpak.com/s/failed-share"},
		}, check: func(_ string, args []driver.NamedValue) {
			foundAlias := false
			for _, arg := range args {
				foundAlias = foundAlias || arg.Value == magnetblock.Aliases(testMagnetHash)[1]
			}
			if !foundAlias {
				t.Fatal("historical lookup omitted base32 hashes")
			}
		}},
		sqlStep{query: "pikpak_shared_link", columns: []string{"share_id", "file_id", "master_user_id", "worker_user_id"}, rows: [][]driver.Value{
			{"old-share", "file-a", "master-a", "worker-a"}, {"failed-share", "file-z", "master-z", "worker-z"},
		}},
		sqlStep{query: "pikpak_shared_link", columns: []string{"share_id", "worker_user_id"}, rows: [][]driver.Value{{"old-share", "worker-a"}}},
		sqlStep{query: "keepshare_shared_link", columns: []string{"auto_id", "state", "host_shared_link"}, rows: [][]driver.Value{
			{int64(11), "BLOCKED", "https://mypikpak.com/s/old-share?act=play"},
			{int64(12), "OK", "https://mypikpak.com/s/old-share"},
		}},
		sqlStep{exec: "UPDATE `keepshare_shared_link`"},
		sqlStep{exec: "UPDATE `keepshare_shared_link`"},
		sqlStep{exec: "DELETE FROM `pikpak_shared_link`", check: func(_ string, args []driver.NamedValue) {
			for _, arg := range args {
				if arg.Value == "worker-z" || arg.Value == "failed-share" {
					t.Fatal("failed native revocation lost its retry mapping")
				}
			}
		}},
	)
	api := testAPI(db, "worker-a", "worker-z")
	var cacheKeys []string
	api.Redis = redis.NewClient(&redis.Options{Addr: "unused.invalid:0"})
	api.Redis.AddHook(cacheDeleteHook{run: func(cmd redis.Cmder) error {
		for _, key := range cmd.Args()[1:] {
			cacheKeys = append(cacheKeys, key.(string))
		}
		cmd.(*redis.IntCmd).SetVal(2)
		return nil
	}})
	var requests int
	stubPikPak(t, func(r *http.Request) (int, string) {
		requests++
		if r.URL.Path != "/drive/v1/share:batchDelete" {
			t.Fatalf("revocation touched cloud files: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "Bearer fake-token-worker-z" {
			return 503, ""
		}
		return 204, ""
	})
	result, err := api.RevokeMagnetShares(context.Background(), "magnet:?xt=urn:btih:"+testMagnetHash)
	if err == nil || result.Revoked != 1 || len(result.Failures) != 1 || result.Failures[0].WorkerUserID != "worker-z" || result.Failures[0].ShareID != "failed-share" {
		t.Fatalf("RevokeMagnetShares() = %+v, %v", result, err)
	}
	if requests != 2 || strings.Join(cacheKeys, ",") != "status:11,status:12" {
		t.Fatalf("requests = %d, deleted cache keys = %v", requests, cacheKeys)
	}
}

func TestRevokeMagnetSharesKeepsMappingWhenCacheCleanupFails(t *testing.T) {
	db := newScriptDB(t,
		sqlStep{query: "JOIN pikpak_file", columns: []string{"share_id", "worker_user_id"}, rows: [][]driver.Value{{"old-share", "worker"}}},
		sqlStep{query: "keepshare_shared_link", columns: []string{"host_shared_link"}},
		sqlStep{query: "pikpak_shared_link", columns: []string{"share_id", "worker_user_id"}, rows: [][]driver.Value{{"old-share", "worker"}}},
		sqlStep{query: "keepshare_shared_link", columns: []string{"auto_id", "host_shared_link"}, rows: [][]driver.Value{{int64(11), "https://mypikpak.com/s/old-share"}}},
	)
	api := testAPI(db, "worker")
	api.Redis = redis.NewClient(&redis.Options{Addr: "unused.invalid:0"})
	api.Redis.AddHook(cacheDeleteHook{run: func(redis.Cmder) error { return errors.New("cache unavailable") }})
	stubPikPak(t, func(r *http.Request) (int, string) {
		if r.URL.Path != "/drive/v1/share:batchDelete" {
			t.Fatalf("unexpected cloud operation %s", r.URL.Path)
		}
		return 204, ""
	})
	result, err := api.RevokeMagnetShares(context.Background(), "magnet:?xt=urn:btih:"+testMagnetHash)
	if err == nil || result.Revoked != 1 || len(result.Failures) != 1 || result.Failures[0].Stage != "cache" {
		t.Fatalf("RevokeMagnetShares() = %+v, %v", result, err)
	}
}

func testAPI(db *gorm.DB, workers ...string) *API {
	api := &API{Dependencies: &hosts.Dependencies{Mysql: db}, cache: freecache.NewCache(1024 * 1024)}
	if db != nil {
		api.q = query.Use(db)
	}
	for _, worker := range workers {
		api.cache.Set([]byte("token:"+worker), []byte("fake-token-"+worker), 3600)
	}
	return api
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func stubPikPak(t *testing.T, handle func(*http.Request) (int, string)) {
	t.Helper()
	previous := resCli
	resCli = resty.New().SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		status, body := handle(r)
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	}))
	t.Cleanup(func() { resCli = previous })
}

// Scripted SQL and Redis hooks keep these behavior tests entirely offline.
type sqlStep struct {
	query   string
	exec    string
	columns []string
	rows    [][]driver.Value
	err     error
	check   func(string, []driver.NamedValue)
}

type sqlScript struct {
	t     *testing.T
	steps []sqlStep
}

func (s *sqlScript) next(query string, args []driver.NamedValue, exec bool) sqlStep {
	s.t.Helper()
	if len(s.steps) == 0 {
		s.t.Fatalf("unexpected SQL: %s", query)
	}
	step := s.steps[0]
	s.steps = s.steps[1:]
	want := step.query
	if exec {
		want = step.exec
	}
	if want == "" || !strings.Contains(query, want) {
		s.t.Fatalf("SQL = %s; want %q (exec=%v)", query, want, exec)
	}
	if step.check != nil {
		step.check(query, args)
	}
	return step
}

func newScriptDB(t *testing.T, steps ...sqlStep) *gorm.DB {
	t.Helper()
	script := &sqlScript{t: t, steps: steps}
	conn := sql.OpenDB(scriptConnector{script})
	conn.SetMaxOpenConns(1)
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn.Close()
		if len(script.steps) != 0 {
			t.Errorf("%d expected SQL operations were not exercised", len(script.steps))
		}
	})
	return db
}

type scriptConnector struct{ script *sqlScript }

func (c scriptConnector) Connect(context.Context) (driver.Conn, error) {
	return scriptConn{c.script}, nil
}
func (c scriptConnector) Driver() driver.Driver { return scriptDriver{} }

type scriptDriver struct{}

func (scriptDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use the scripted connector")
}

type scriptConn struct{ script *sqlScript }

func (scriptConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepared SQL")
}
func (scriptConn) Close() error              { return nil }
func (scriptConn) Begin() (driver.Tx, error) { return scriptTx{}, nil }
func (scriptConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return scriptTx{}, nil
}
func (c scriptConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	step := c.script.next(query, args, false)
	return &scriptRows{columns: step.columns, rows: step.rows}, step.err
}
func (c scriptConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	step := c.script.next(query, args, true)
	return scriptResult{}, step.err
}

type scriptResult struct{}

func (scriptResult) LastInsertId() (int64, error) { return 1, nil }
func (scriptResult) RowsAffected() (int64, error) { return 1, nil }

type scriptTx struct{}

func (scriptTx) Commit() error   { return nil }
func (scriptTx) Rollback() error { return nil }

type scriptRows struct {
	columns []string
	rows    [][]driver.Value
}

func (r *scriptRows) Columns() []string { return r.columns }
func (r *scriptRows) Close() error      { return nil }
func (r *scriptRows) Next(dest []driver.Value) error {
	if len(r.rows) == 0 {
		return io.EOF
	}
	copy(dest, r.rows[0])
	r.rows = r.rows[1:]
	return nil
}

type cacheDeleteHook struct{ run func(redis.Cmder) error }

func (cacheDeleteHook) DialHook(redis.DialHook) redis.DialHook {
	return func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("unexpected Redis network access")
	}
}
func (h cacheDeleteHook) ProcessHook(redis.ProcessHook) redis.ProcessHook {
	return func(_ context.Context, cmd redis.Cmder) error { return h.run(cmd) }
}
func (cacheDeleteHook) ProcessPipelineHook(redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(context.Context, []redis.Cmder) error { return errors.New("unexpected Redis pipeline") }
}

var _ driver.QueryerContext = scriptConn{}
var _ driver.ExecerContext = scriptConn{}
var _ driver.ConnBeginTx = scriptConn{}
