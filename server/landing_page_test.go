// Copyright 2026 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestLandingPageDocumentVisitsMySQL(t *testing.T) {
	db := newLinkVisitsTestDB(t)
	seedLinkVisitTestRecord(t, db, 1)
	seedLinkVisitTestRecord(t, db, 2)
	router := gin.New()
	router.Use(recordLandingPageVisit(db))
	router.Any("/*path", func(c *gin.Context) {
		switch c.Query("response") {
		case "redirect":
			c.Redirect(http.StatusFound, "/console/shared/wsl-status?id=1")
		case "error":
			c.Status(http.StatusServiceUnavailable)
		default:
			c.Data(http.StatusOK, "text/html", []byte("<html>Landing page</html>"))
		}
	})

	for _, visit := range []struct {
		method string
		url    string
		pv     int64
		ips    int64
	}{
		{http.MethodGet, "/console/shared/status?id=1", 1, 1},
		{http.MethodGet, "/console/shared/status?id=1", 2, 1},
		{http.MethodGet, "/console/shared/wsl-status?id=1", 3, 1},
		{http.MethodGet, "/console/shared/wsl-status?id=2", 3, 1},
		{http.MethodGet, "/api/shared_link?id=1", 3, 1},
		{http.MethodGet, "/console/assets/main.js?id=1", 3, 1},
		{http.MethodGet, "/console/shared-links?id=1", 3, 1},
		{http.MethodGet, "/abcd1234/magnet:?xt=example", 3, 1},
		{http.MethodHead, "/console/shared/status?id=1", 3, 1},
		{http.MethodGet, "/console/shared/status?id=1&response=redirect", 3, 1},
		{http.MethodGet, "/console/shared/status?id=1&response=error", 3, 1},
		{http.MethodGet, "/console/shared/status?id=0", 3, 1},
		{http.MethodGet, "/console/shared/status?id=-1", 3, 1},
		{http.MethodGet, "/console/shared/status?id=invalid", 3, 1},
		{http.MethodGet, "/console/shared/status?id=9223372036854775808", 3, 1},
		{http.MethodGet, "/console/shared/status?id=999", 3, 1},
	} {
		req := httptest.NewRequest(visit.method, visit.url, nil)
		req.RemoteAddr = "192.0.2.1:45000"
		req.Header.Set("X-Forwarded-For", "198.51.100.2")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assertLinkVisitCounts(t, db, 1, visit.pv, visit.ips)
		if visit.method == http.MethodGet && isLandingPage(req.URL.Path) && response.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("%s: landing documents must not be cached", visit.url)
		}
	}
	assertLinkVisitCounts(t, db, 2, 1, 1)

	for i, visit := range []struct {
		forwardedFor string
		realIP       string
		ips          int64
	}{
		{"198.51.100.3, 127.0.0.1", "", 2},
		{"198.51.100.4, 127.0.0.1", "", 3},
		{"198.51.100.3, 127.0.0.1", "", 3},
		{"", "", 4},
		{"", "198.51.100.5", 5},
		{"198.51.100.3", "198.51.100.6", 5},
		{"invalid", "198.51.100.5", 5},
		{"invalid", "", 5},
	} {
		req := httptest.NewRequest(http.MethodGet, "/console/shared/status?id=1", nil)
		req.RemoteAddr = "192.0.2.1:45000"
		req.Header.Set("X-Forwarded-For", visit.forwardedFor)
		req.Header.Set("X-Real-IP", visit.realIP)
		router.ServeHTTP(httptest.NewRecorder(), req)
		assertLinkVisitCounts(t, db, 1, int64(4+i), visit.ips)
	}

	if err := db.Exec("DROP TABLE keepshare_shared_link_visitor_ip").Error; err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/console/shared/status?id=1", nil)
	req.RemoteAddr = "192.0.2.2:45000"
	router.ServeHTTP(response, req)
	if response.Code != http.StatusOK || response.Body.String() != "<html>Landing page</html>" {
		t.Error("a statistics failure changed the landing page response")
	}
	var record model.SharedLink
	if err := db.First(&record, 1).Error; err != nil {
		t.Fatal(err)
	}
	if record.PageViews != 11 || record.VisitorIps != 5 {
		t.Error("a failed statistics write partially changed the counters")
	}
}

func TestProxiedLandingPageVisitsMySQL(t *testing.T) {
	db := newLinkVisitsTestDB(t)
	seedLinkVisitTestRecord(t, db, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>Proxied landing page</html>"))
	}))
	defer upstream.Close()
	old := viper.Get("console_proxy_url")
	viper.Set("console_proxy_url", upstream.URL)
	t.Cleanup(func() { viper.Set("console_proxy_url", old) })
	router := gin.New()
	router.GET("/console/*path", recordLandingPageVisit(db), consoleProxy())
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/console/shared/wsl-status?id=1", nil)
	req.RemoteAddr = "192.0.2.1:45000"
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	req = req.WithContext(ctx)
	router.ServeHTTP(response, req)
	if response.Code != http.StatusOK || response.Body.String() != "<html>Proxied landing page</html>" {
		t.Errorf("unexpected proxied response: %d %s", response.Code, response.Body.String())
	}
	if values := response.Header().Values("Cache-Control"); len(values) != 1 || values[0] != "no-store" {
		t.Errorf("expected one non-cacheable policy, got %v", values)
	}
	assertLinkVisitCounts(t, db, 1, 1, 1)
}
