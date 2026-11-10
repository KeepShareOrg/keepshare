// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package link

import (
	"crypto/sha1"
	"fmt"
	"testing"
)

const (
	ed2kHash  = "8549fd92d11e47e3d9c3379e5ec8f842"
	ed2kRaw   = "ed2k://|file|The_Matrix.avi|735934464|8549FD92D11E47E3D9C3379E5EC8F842|/"
	ed2kPlain = "ed2k://|file|The_Matrix.avi|735934464|" + ed2kHash + "|/"
)

func TestParseEd2k(t *testing.T) {
	for name, tc := range map[string]struct {
		raw       string
		wantOK    bool
		wantName  string
		wantSize  int64
		wantHash  string
		wantCanon string
	}{
		"uppercase hash": {
			raw:       ed2kRaw,
			wantOK:    true,
			wantName:  "The_Matrix.avi",
			wantSize:  735934464,
			wantHash:  ed2kHash,
			wantCanon: ed2kPlain,
		},
		"uppercase scheme": {
			raw:       "ED2K://|file|a.bin|1024|" + ed2kHash + "|/",
			wantOK:    true,
			wantName:  "a.bin",
			wantSize:  1024,
			wantHash:  ed2kHash,
			wantCanon: "ed2k://|file|a.bin|1024|" + ed2kHash + "|/",
		},
		"aich hash is dropped from the canonical form": {
			raw:       "ed2k://|file|a.bin|1024|" + ed2kHash + "|h=ABCDEFGHIJKLMNOPQRSTUVWXYZ234567|/",
			wantOK:    true,
			wantName:  "a.bin",
			wantSize:  1024,
			wantHash:  ed2kHash,
			wantCanon: "ed2k://|file|a.bin|1024|" + ed2kHash + "|/",
		},
		"surrounding spaces": {
			raw:       "  " + ed2kPlain + "  ",
			wantOK:    true,
			wantName:  "The_Matrix.avi",
			wantSize:  735934464,
			wantHash:  ed2kHash,
			wantCanon: ed2kPlain,
		},
		"server link": {
			raw: "ed2k://|server|1.2.3.4|4661|/",
		},
		"serverlist link": {
			raw: "ed2k://|serverlist|https://example.invalid/server.met|/",
		},
		"missing hash": {
			raw: "ed2k://|file|a.bin|1024||/",
		},
		"truncated hash": {
			raw: "ed2k://|file|a.bin|1024|8549fd92|/",
		},
		"non-hex hash": {
			raw: "ed2k://|file|a.bin|1024|zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz|/",
		},
		"missing final delimiter": {
			raw: "ed2k://|file|a.bin|1024|" + ed2kHash,
		},
		"wrong final delimiter": {
			raw: "ed2k://|file|a.bin|1024|" + ed2kHash + "|garbage",
		},
		"extra trailing field": {
			raw: "ed2k://|file|a.bin|1024|" + ed2kHash + "|/|junk",
		},
		"empty name": {
			raw: "ed2k://|file||1024|" + ed2kHash + "|/",
		},
		"negative size": {
			raw: "ed2k://|file|a.bin|-1|" + ed2kHash + "|/",
		},
		"not a number size": {
			raw: "ed2k://|file|a.bin|abc|" + ed2kHash + "|/",
		},
		"magnet link": {
			raw: "magnet:?xt=urn:btih:" + ed2kHash,
		},
		"http link": {
			raw: "https://example.invalid/file.bin",
		},
		"empty": {},
	} {
		t.Run(name, func(t *testing.T) {
			f, ok := ParseEd2k(tc.raw)
			if ok != tc.wantOK {
				t.Fatalf("ParseEd2k(%q) ok = %v, want %v", tc.raw, ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if f.Name != tc.wantName || f.Size != tc.wantSize || f.Hash != tc.wantHash {
				t.Fatalf("ParseEd2k(%q) = %+v, want name %q size %d hash %q", tc.raw, f, tc.wantName, tc.wantSize, tc.wantHash)
			}
			if got := f.String(); got != tc.wantCanon {
				t.Fatalf("canonical = %q, want %q", got, tc.wantCanon)
			}
		})
	}
}

func TestHash(t *testing.T) {
	httpLink := "https://example.invalid/file.bin"
	httpSHA1 := fmt.Sprintf("%x", sha1.Sum([]byte(httpLink)))
	for name, tc := range map[string]struct {
		raw  string
		want string
	}{
		"ed2k uses the file hash":             {raw: ed2kRaw, want: ed2kHash},
		"invalid ed2k has no hash":            {raw: "ed2k://|server|1.2.3.4|4661|/", want: ""},
		"ed2k without terminator has no hash": {raw: "ed2k://|file|a.bin|1024|" + ed2kHash + "|garbage", want: ""},
		"magnet uses the info hash":           {raw: "magnet:?xt=urn:btih:" + ed2kHash, want: ed2kHash},
		"http uses the sha1 of the link":      {raw: httpLink, want: httpSHA1},
	} {
		t.Run(name, func(t *testing.T) {
			if got := Hash(tc.raw); got != tc.want {
				t.Fatalf("Hash(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestSimplifyEd2k(t *testing.T) {
	raw := "ed2k://|file|a.bin|1024|" + ed2kHash + "|h=ABCDEFGHIJKLMNOPQRSTUVWXYZ234567|/"
	if got := Simplify(raw); got != "ed2k://|file|a.bin|1024|"+ed2kHash+"|/" {
		t.Fatalf("Simplify(%q) = %q", raw, got)
	}
	if got := Simplify("ed2k://|server|1.2.3.4|4661|/"); got != "ed2k://|server|1.2.3.4|4661|/" {
		t.Fatalf("invalid ed2k must be returned unchanged, got %q", got)
	}
}
