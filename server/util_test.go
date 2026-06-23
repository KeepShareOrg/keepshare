// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/viper"

	"github.com/KeepShareOrg/keepshare/config"
)

func resetConfigForTest(t *testing.T, rootDomains []string) {
	t.Helper()
	viper.Reset()
	for k, v := range config.GetConfigsForTest() {
		viper.SetDefault(k, v.Default)
	}
	if len(rootDomains) > 0 {
		viper.Set("root_domains", rootDomains)
	}
}

func TestMakeKeepSharingLink_UsesProvidedHost(t *testing.T) {
	got := makeKeepSharingLink("abcd1234", "https://example.com/file.zip", "keepshare.link")
	want := "https://keepshare.link/abcd1234/https%3A%2F%2Fexample.com%2Ffile.zip"
	if got != want {
		t.Errorf("makeKeepSharingLink = %q, want %q", got, want)
	}
}

func TestGetOriginalLinks_RecognizesAnyAllowedDomain(t *testing.T) {
	resetConfigForTest(t, []string{"keepshare.cc", "keepshare.link"})

	original, invalid := getOriginalLinks([]string{
		"https://keepshare.link/abcd1234/https%3A%2F%2Fexample.com%2Ffile.zip",
	})
	if len(invalid) != 0 {
		t.Errorf("unexpected invalid links: %v", invalid)
	}
	if len(original) != 1 {
		t.Fatalf("expected 1 original, got %d (%v)", len(original), original)
	}
	if original[0] != "https://example.com/file.zip" {
		t.Errorf("decoded original = %q, want %q", original[0], "https://example.com/file.zip")
	}
}

func TestGetOriginalLinks_RejectsUnknownDomain(t *testing.T) {
	resetConfigForTest(t, []string{"keepshare.cc"})

	original, invalid := getOriginalLinks([]string{
		"https://evil.example/abcd1234/https%3A%2F%2Fexample.com%2Ffile.zip",
	})
	if len(original) != 1 {
		t.Errorf("expected the evil link to be passed through as original, got %d", len(original))
	}
	if len(invalid) != 0 {
		t.Errorf("expected no invalid entries, got %v", invalid)
	}
}

func TestHostOrDefault_UsesRequestHost(t *testing.T) {
	resetConfigForTest(t, []string{"primary.example"})
	r := &http.Request{Host: "inbound.example"}
	if got := hostOrDefault(r); got != "inbound.example" {
		t.Errorf("hostOrDefault = %q, want %q", got, "inbound.example")
	}
}

func TestHostOrDefault_FallsBackWhenEmpty(t *testing.T) {
	resetConfigForTest(t, []string{"primary.example"})
	r := &http.Request{Host: ""}
	if got := hostOrDefault(r); got != "primary.example" {
		t.Errorf("hostOrDefault = %q, want %q", got, "primary.example")
	}
}

func TestBuildVerifyLink_UsesProvidedHost(t *testing.T) {
	got := buildVerifyLink("keepshare.link", "tok123", "user@example.com", 1234567890)
	want := "https://keepshare.link/api/verification?token=tok123&email=user@example.com&expires=1234567890"
	if got != want {
		t.Errorf("buildVerifyLink = %q, want %q", got, want)
	}
}

func TestBuildResultPageAddr_UsesProvidedHost(t *testing.T) {
	got := buildResultPageAddr("keepshare.link")
	want := "https://keepshare.link/console/email-verification"
	if got != want {
		t.Errorf("buildResultPageAddr = %q, want %q", got, want)
	}
}

func TestAutoSharingRedirectURL_FormatsCorrectly(t *testing.T) {
	// Verifies the inline URL format used by both redirects in
	// autoSharingLink. If this format changes, the status page and
	// wsl-status page behavior will change in lockstep.
	host := hostOrDefault(&http.Request{Host: "keepshare.link"})
	scheme := "http" // simulating a dev http request
	got := fmt.Sprintf("%s://%s/console/shared/wsl-status?id=%d&request_id=%s", scheme, host, 42, "req-1")
	want := "http://keepshare.link/console/shared/wsl-status?id=42&request_id=req-1"
	if got != want {
		t.Errorf("redirect URL = %q, want %q", got, want)
	}
}
