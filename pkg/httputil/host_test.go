// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package httputil

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestRequestHost_PlainHost(t *testing.T) {
	r := &http.Request{Host: "keepshare.cc"}
	if got := RequestHost(r); got != "keepshare.cc" {
		t.Errorf("RequestHost = %q, want %q", got, "keepshare.cc")
	}
}

func TestRequestHost_StripsPort(t *testing.T) {
	r := &http.Request{Host: "keepshare.cc:443"}
	if got := RequestHost(r); got != "keepshare.cc" {
		t.Errorf("RequestHost = %q, want %q", got, "keepshare.cc")
	}
}

func TestRequestHost_StripsPortForCustomPort(t *testing.T) {
	r := &http.Request{Host: "localhost:8080"}
	if got := RequestHost(r); got != "localhost" {
		t.Errorf("RequestHost = %q, want %q", got, "localhost")
	}
}

func TestRequestHost_IPv6WithPort(t *testing.T) {
	r := &http.Request{Host: "[::1]:8080"}
	if got := RequestHost(r); got != "::1" {
		t.Errorf("RequestHost = %q, want %q", got, "::1")
	}
}

func TestRequestHost_IPv6BracketedNoPort(t *testing.T) {
	r := &http.Request{Host: "[::1]"}
	if got := RequestHost(r); got != "::1" {
		t.Errorf("RequestHost = %q, want %q", got, "::1")
	}
}

func TestRequestHost_XForwardedHostWins(t *testing.T) {
	r := &http.Request{
		Host: "internal:8080",
		Header: http.Header{
			"X-Forwarded-Host": []string{"keepshare.link"},
		},
	}
	if got := RequestHost(r); got != "keepshare.link" {
		t.Errorf("RequestHost = %q, want %q", got, "keepshare.link")
	}
}

func TestRequestHost_Lowercases(t *testing.T) {
	r := &http.Request{Host: "KeepShare.CC"}
	if got := RequestHost(r); got != "keepshare.cc" {
		t.Errorf("RequestHost = %q, want %q", got, "keepshare.cc")
	}
}

func TestRequestHost_EmptyHostReturnsEmpty(t *testing.T) {
	r := &http.Request{Host: ""}
	if got := RequestHost(r); got != "" {
		t.Errorf("RequestHost = %q, want empty", got)
	}
}

func TestRequestScheme_FromTLS(t *testing.T) {
	r := &http.Request{TLS: &tls.ConnectionState{}}
	if got := RequestScheme(r); got != "https" {
		t.Errorf("RequestScheme = %q, want %q", got, "https")
	}
}

func TestRequestScheme_DefaultsToHTTP(t *testing.T) {
	r := &http.Request{}
	if got := RequestScheme(r); got != "http" {
		t.Errorf("RequestScheme = %q, want %q", got, "http")
	}
}

func TestRequestScheme_XForwardedProtoWins(t *testing.T) {
	r := &http.Request{
		Header: http.Header{
			"X-Forwarded-Proto": []string{"https"},
		},
	}
	if got := RequestScheme(r); got != "https" {
		t.Errorf("RequestScheme = %q, want %q", got, "https")
	}
}

func TestRequestScheme_XForwardedProtoHTTP(t *testing.T) {
	r := &http.Request{
		Header: http.Header{
			"X-Forwarded-Proto": []string{"http"},
		},
	}
	if got := RequestScheme(r); got != "http" {
		t.Errorf("RequestScheme = %q, want %q", got, "http")
	}
}
