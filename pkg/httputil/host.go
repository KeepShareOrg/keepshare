// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package httputil provides small HTTP helpers used across the server layer.
package httputil

import (
	"net"
	"net/http"
	"strings"
)

// RequestHost returns the inbound request's host with any port stripped.
// It honors the X-Forwarded-Host header when present (assumes a trusted
// reverse proxy in front of the service). The result is lowercased and
// whitespace-trimmed. Returns "" if no host is available.
func RequestHost(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	// net.SplitHostPort handles bracketed IPv6 hosts (e.g. "[::1]:8080")
	// correctly. If there's no port, it returns an error and we strip any
	// surrounding brackets for bracketed-IPv6-without-port values.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	} else if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	return strings.ToLower(strings.TrimSpace(host))
}

// RequestScheme returns "https" or "http" for the inbound request.
// X-Forwarded-Proto takes precedence; otherwise checks r.TLS; defaults
// to "http" when neither indicates an encrypted connection.
func RequestScheme(r *http.Request) string {
	if p := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))); p == "https" || p == "http" {
		return p
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
