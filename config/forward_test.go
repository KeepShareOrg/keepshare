// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestForwardMailsDefaults(t *testing.T) {
	resetViper(t)
	if got := ForwardMailsEnable(); got != false {
		t.Fatalf("default enable: want false, got %v", got)
	}
	if got := ForwardMailsMonitorURL(); got != "" {
		t.Fatalf("default monitor_url: want '', got %q", got)
	}
	if got := ForwardMailsNotice(); got == "" {
		t.Fatalf("default notice should be non-empty")
	}
	if got := ForwardMailsAccounts(); len(got) != 0 {
		t.Fatalf("default accounts: want empty, got %d", len(got))
	}
}

func TestForwardMailsFromYAML(t *testing.T) {
	resetViper(t)
	yaml := `forward_mails:
  enable: true
  monitor_url: "wss://demo/api/v1/monitor/messages"
  notice: "NOTICE"
  accounts:
    - username: "a"
      server: "smtp.example.com"
      port: 587
      email: "a@example.com"
      password: "secret"
    - username: "b"
      server: "smtp.example.com"
      port: 465
      email: "b@example.com"
      password: "pw"
`
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(filepath.Join(dir, "config.yaml"))
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	if !ForwardMailsEnable() {
		t.Fatal("expected enable=true from yaml")
	}
	if got := ForwardMailsMonitorURL(); got != "wss://demo/api/v1/monitor/messages" {
		t.Fatalf("monitor_url mismatch: %q", got)
	}
	if got := ForwardMailsNotice(); got != "NOTICE" {
		t.Fatalf("notice mismatch: %q", got)
	}
	accs := ForwardMailsAccounts()
	if len(accs) != 2 {
		t.Fatalf("want 2 accounts, got %d", len(accs))
	}
	if accs[0].Email != "a@example.com" || accs[0].Port != 587 {
		t.Fatalf("accs[0] wrong: %+v", accs[0])
	}
}

func TestForwardMails(t *testing.T) {
	resetViper(t)
	viper.Set("forward_mails.enable", true)
	cfg := ForwardMails()
	if !cfg.Enable || cfg.MonitorURL != "" || cfg.Notice == "" || len(cfg.Accounts) != 0 {
		t.Fatalf("aggregate getter wrong: %+v", cfg)
	}
}
