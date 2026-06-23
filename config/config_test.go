// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"testing"

	"github.com/spf13/viper"
)

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	for k, v := range configs {
		viper.SetDefault(k, v.Default)
	}
}

func TestRootDomains_LegacySingleValue(t *testing.T) {
	resetViper(t)
	viper.Set("root_domain", "keepshare.cc")

	got := RootDomains()
	if len(got) != 1 || got[0] != "keepshare.cc" {
		t.Errorf("RootDomains() = %v, want [keepshare.cc]", got)
	}
}

func TestRootDomains_MultiValueList(t *testing.T) {
	resetViper(t)
	viper.Set("root_domains", []string{"keepshare.cc", "keepshare.link"})

	got := RootDomains()
	if len(got) != 2 || got[0] != "keepshare.cc" || got[1] != "keepshare.link" {
		t.Errorf("RootDomains() = %v, want [keepshare.cc keepshare.link]", got)
	}
}

func TestRootDomains_BothSet_ListWins(t *testing.T) {
	resetViper(t)
	viper.Set("root_domain", "legacy.example")
	viper.Set("root_domains", []string{"a.example", "b.example"})

	got := RootDomains()
	if len(got) != 2 || got[0] != "a.example" || got[1] != "b.example" {
		t.Errorf("RootDomains() = %v, want [a.example b.example]", got)
	}
}

func TestIsRootDomain_Match(t *testing.T) {
	resetViper(t)
	viper.Set("root_domains", []string{"keepshare.cc", "keepshare.link"})

	if !IsRootDomain("keepshare.cc") {
		t.Error("IsRootDomain(keepshare.cc) = false, want true")
	}
	if !IsRootDomain("keepshare.link") {
		t.Error("IsRootDomain(keepshare.link) = false, want true")
	}
}

func TestIsRootDomain_CaseInsensitive(t *testing.T) {
	resetViper(t)
	viper.Set("root_domains", []string{"keepshare.cc"})

	if !IsRootDomain("KeepShare.CC") {
		t.Error("IsRootDomain(KeepShare.CC) = false, want true")
	}
}

func TestIsRootDomain_NoMatch(t *testing.T) {
	resetViper(t)
	viper.Set("root_domains", []string{"keepshare.cc"})

	if IsRootDomain("evil.example") {
		t.Error("IsRootDomain(evil.example) = true, want false")
	}
}

func TestRootDomain_ReturnsFirstElement(t *testing.T) {
	resetViper(t)
	viper.Set("root_domains", []string{"first.example", "second.example"})

	if got := RootDomain(); got != "first.example" {
		t.Errorf("RootDomain() = %q, want %q", got, "first.example")
	}
}
