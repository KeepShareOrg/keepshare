// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	callerPretty = func(caller *runtime.Frame) (function string, file string) {
		dir, name := filepath.Split(caller.File)
		file = fmt.Sprintf("%s/%s:%d", filepath.Base(dir), name, caller.Line)
		_, function = filepath.Split(caller.Function)
		return function, file
	}

	// JSONLogFormatter formats logs into parsable json.
	JSONLogFormatter = &log.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000-07:00", CallerPrettyfier: callerPretty, DisableHTMLEscape: true}

	// TextLogFormatter formats logs into text.
	TextLogFormatter = &log.TextFormatter{TimestampFormat: "2006-01-02T15:04:05.000-07:00", FullTimestamp: true, CallerPrettyfier: callerPretty}
)

func initLogger() {
	level, err := log.ParseLevel(LogLevel())
	if err == nil {
		log.SetLevel(level)
	}

	log.SetOutput(LogWriter(LogOutput()))
	JSONLogFormatter.PrettyPrint = LogPretty()

	switch LogFormat() {
	case "text", "txt":
		log.SetFormatter(TextLogFormatter)
	default:
		log.SetFormatter(JSONLogFormatter)
	}
}

// LogWriter get an io.Writer according to the output.
func LogWriter(output string) io.Writer {
	switch output {
	case "", "stdout", "/dev/stdout":
		return os.Stdout
	case "stderr", "/dev/stderr":
		return os.Stderr
	default:
		return &lumberjack.Logger{
			Filename:   output,
			MaxSize:    200 * 1024 * 1024, // 200MB
			MaxAge:     31,                // 31 days
			MaxBackups: 0,
			LocalTime:  true,
			Compress:   false,
		}
	}
}
