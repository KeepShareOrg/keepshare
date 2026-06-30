// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import "github.com/spf13/viper"

// ForwardAccount is one SMTP sender account in the rotating pool.
type ForwardAccount struct {
	Username string
	Server   string
	Port     int
	Email    string
	Password string
}

// ForwardMailsConfig is the resolved forward_mails configuration block.
type ForwardMailsConfig struct {
	Enable     bool
	MonitorURL string
	Notice     string
	Accounts   []ForwardAccount
}

// ForwardMailsEnable returns the master switch.
func ForwardMailsEnable() bool { return viper.GetBool("forward_mails.enable") }

// ForwardMailsMonitorURL returns the configured WS endpoint, or empty to use the
// mail_server fallback.
func ForwardMailsMonitorURL() string { return viper.GetString("forward_mails.monitor_url") }

// ForwardMailsNotice returns the prefix inserted into every forwarded body.
func ForwardMailsNotice() string { return viper.GetString("forward_mails.notice") }

// ForwardMailsAccounts reads forward_mails.accounts and decodes each element.
func ForwardMailsAccounts() []ForwardAccount {
	raw := viper.Get("forward_mails.accounts")
	list, _ := raw.([]any)
	if list == nil {
		return nil
	}
	out := make([]ForwardAccount, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		port, _ := m["port"].(int)
		out = append(out, ForwardAccount{
			Username: asString(m["username"]),
			Server:   asString(m["server"]),
			Port:     port,
			Email:    asString(m["email"]),
			Password: asString(m["password"]),
		})
	}
	return out
}

// ForwardMails aggregates the individual getters into a struct.
func ForwardMails() ForwardMailsConfig {
	return ForwardMailsConfig{
		Enable:     ForwardMailsEnable(),
		MonitorURL: ForwardMailsMonitorURL(),
		Notice:     ForwardMailsNotice(),
		Accounts:   ForwardMailsAccounts(),
	}
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
