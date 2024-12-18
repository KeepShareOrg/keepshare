// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/spf13/viper"
)

// define main configs.
var (
	DefaultHost = func() string { return viper.GetString("host_default") }
	RootDomain  = func() string { return viper.GetString("root_domain") }
	ListenHTTP  = func() string { return viper.GetString("listen_http") }
	ListenHTTPS = func() string { return viper.GetString("listen_https") }

	LogLevel        = func() string { return viper.GetString("log_level") }
	LogFormat       = func() string { return viper.GetString("log_format") }
	LogOutput       = func() string { return viper.GetString("log_output") }
	LogPretty       = func() bool { return viper.GetBool("log_pretty") }
	LogBackups      = func() int { return viper.GetInt("log_backups") }
	LogAgeInDay     = func() int { return viper.GetInt("log_age_in_day") }
	AccessLogOutput = func() string { return viper.GetString("log_access_output") }

	GoogleRecaptchaSecret = func() string { return viper.GetString("google_recaptcha_secret") }

	ConsoleProxyURL = func() string { return viper.GetString("console_proxy_url") }

	dbMySQL    = func() string { return viper.GetString("db_mysql") }
	dbRedis    = func() string { return viper.GetString("db_redis") }
	mailServer = func() string { return viper.GetString("mail_server") }
)

var configs = map[string]properties{
	"root_domain":  {"localhost", "Domain for this project, including web pages or keep sharing links"},
	"host_default": {"pikpak", "When no host is specified, this host is used by default"},
	"listen_http":  {":8080", "HTTP server listen address"},
	"listen_https": {"", "HTTPS server listen address"},

	"log_level":         {"info", "Options: panic, fatal, error, warn, info, debug, trace"},
	"log_format":        {"json", "Options: json, text"},
	"log_output":        {"", "The log output, default to stdout"},
	"log_pretty":        {false, "Print indented json logs if this value is true and log_format is json"},
	"log_backups":       {10, "The maximum number of old log files to retain, only take effect when the log_output is set to a file"},
	"log_age_in_day":    {7, "The maximum age of old log files to retain, only take effect when the log_output is set to a file"},
	"log_access_output": {"", "The access log output, default same to log_output"},

	"google_recaptcha_secret": {"", "The google reCAPTCHA secret key"},

	"db_mysql":    {"user:password@(127.0.0.1:3306)/keepshare?parseTime=True&loc=Local", "Mysql dsn"},
	"db_redis":    {"redis://localhost:6379?dial_timeout=2s&read_timeout=2s&max_retries=2", "Redis url"},
	"mail_server": {"http://localhost", "Mail server to receive and send emails"},

	"console_proxy_url": {"", "If not empty, all the `/console/*` requests will be proxy to this url, mainly used for local testing."},
}

type properties struct {
	Default     any
	Description string
}

const envPrefix = "KS"

func init() {
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	for k, v := range configs {
		viper.SetDefault(k, v.Default)
	}
}

// Load all configs.
func Load() error {
	if err := loadConfig(); err != nil {
		return err
	}

	log.SetLevel(LogLevel())
	log.SetFormatter(LogFormat(), LogPretty())
	log.SetOutput(LogOutput(), &log.OutputOptions{
		MaxSizeInMB: 1024,
		MaxAgeInDay: LogAgeInDay(),
		MaxBackups:  LogBackups(),
	})

	if err := initMysql(); err != nil {
		return err
	}

	if err := initRedis(); err != nil {
		return err
	}

	if err := initMail(); err != nil {
		return err
	}

	return nil
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/keepshare/")
	viper.AddConfigPath("./conf/")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("config file not found")
		} else {
			log.Error("read config err:", err)
			return err
		}
	}
	return nil
}

// Help get help messages.
func Help() string {
	s := &strings.Builder{}
	s.WriteString(`The program will search configuration file in ./conf/ and /etc/keepshare/ for the first file named config.json or config.yaml or config.ini.
It also supports reading from environment variables, non-empty environment variables have a higher priority than files.

Available configurations:
`)
	var keys []string
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := configs[k]
		fmt.Fprintf(s, `
item    : %s
env     : %s
default : %v
desc    : %s
`,
			k, envPrefix+"_"+strings.ReplaceAll(strings.ToUpper(k), ".", "_"), v.Default, v.Description)
	}
	return s.String()
}
