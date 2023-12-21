// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// define aliases to logrus.
var (
	log = logrus.StandardLogger()

	Debug   = log.Debug
	Debugf  = log.Debugf
	Error   = log.Error
	Errorf  = log.Errorf
	Fatal   = log.Fatal
	Fatalf  = log.Fatalf
	Info    = log.Info
	Infof   = log.Infof
	Print   = log.Print
	Printf  = log.Printf
	Println = log.Println
	Trace   = log.Trace
	Tracef  = log.Tracef
	Warn    = log.Warn
	Warnf   = log.Warnf

	WithContext = log.WithContext
	WithError   = log.WithError
	WithField   = log.WithField
	WithFields  = log.WithFields

	RegisterExitHandler = logrus.RegisterExitHandler
)

// define type aliases to logrus.
type (
	Fields = logrus.Fields
	Logger = logrus.Logger
)

func init() {
	log.AddHook(&dataContextHook{})
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
	log.SetFormatter(JSONLogFormatter)
	log.SetLevel(logrus.DebugLevel)
}

var (
	// JSONLogFormatter formats logs into parsable json.
	JSONLogFormatter = &logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000-07:00", CallerPrettyfier: callerPretty, DisableHTMLEscape: true}

	// TextLogFormatter formats logs into text.
	TextLogFormatter = &logrus.TextFormatter{TimestampFormat: "2006-01-02T15:04:05.000-07:00", FullTimestamp: true, CallerPrettyfier: callerPretty}
)

// Log returns the default logger.
func Log() *logrus.Logger {
	return log
}

// New creates a new logger.
func New() *logrus.Logger {
	l := logrus.New()
	l.AddHook(&dataContextHook{})
	return l
}

// SetLevel sets the logger level.
func SetLevel(l string) {
	level, err := logrus.ParseLevel(l)
	if err == nil {
		log.SetLevel(level)
	}
}

// SetOutput sets the logger output.
func SetOutput(filename string, options *OutputOptions) {
	log.SetOutput(Output(filename, options))
}

// SetFormatter sets the standard logger formatter.
func SetFormatter(f string, pretty bool) {
	switch strings.ToLower(f) {
	case "text", "txt":
		logrus.SetFormatter(TextLogFormatter)
	default:
		if pretty {
			JSONLogFormatter.PrettyPrint = pretty
		}
		logrus.SetFormatter(JSONLogFormatter)
	}
}

// Output get an io.Writer according to the filename and options.
func Output(filename string, options *OutputOptions) io.Writer {
	switch filename {
	case "", "stdout", "/dev/stdout":
		return os.Stdout
	case "stderr", "/dev/stderr":
		return os.Stderr
	default:
		l := &lumberjack.Logger{
			Filename:  filename,
			LocalTime: true,
			Compress:  false,
		}
		if options != nil {
			if options.MaxSizeInMB > 0 {
				l.MaxSize = options.MaxSizeInMB
			}
			if options.MaxAgeInDay > 0 {
				l.MaxAge = options.MaxAgeInDay
			}
			if options.MaxBackups > 0 {
				l.MaxBackups = options.MaxBackups
			}
			if options.Rotate != "" {
				autoRotate(l, options.Rotate)
			}
		}
		return l
	}
}

func autoRotate(l *lumberjack.Logger, r string) {
	var timeFormat string
	switch strings.ToLower(r) {
	case "h", "hour":
		timeFormat = "2006010215"
	case "d", "day":
		timeFormat = "20060102"
	default:
		log.Error("invalid log rotate:", r)
		return
	}

	go func() {
		last := time.Now().Format(timeFormat)
		for {
			time.Sleep(time.Second)

			now := time.Now().Format(timeFormat)
			if now != last {
				if err := l.Rotate(); err == nil {
					last = now
				} else {
					log.Errorf("rotate file: %s, err: %v", l.Filename, err)
				}
			}
		}
	}()
}

// OutputOptions set the options to the log output files.
type OutputOptions struct {
	Rotate      string // options: d or day, h or hour
	MaxSizeInMB int
	MaxAgeInDay int
	MaxBackups  int
}

// IsDebugEnabled checks if the log level is greater than the debug level.
func IsDebugEnabled() bool {
	return log.IsLevelEnabled(logrus.DebugLevel)
}

func callerPretty(caller *runtime.Frame) (function string, file string) {
	dir, name := filepath.Split(caller.File)
	file = fmt.Sprintf("%s/%s:%d", filepath.Base(dir), name, caller.Line)
	_, function = filepath.Split(caller.Function)
	return function, file
}
