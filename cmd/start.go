// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/KeepShareOrg/keepshare/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	var startAsDaemon bool
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		Run: func(cmd *cobra.Command, _ []string) {
			start(startAsDaemon)
		},
	}
	cmd.Flags().BoolVarP(&startAsDaemon, "daemon", "d", false, "Run server as a daemon")

	rootCmd.AddCommand(cmd)
}

func start(daemon bool) {
	// check if the server is running
	pid := getPid()
	if pid != -1 {
		_, err := os.FindProcess(pid)
		if err == nil {
			stdLog.Print("server is running, pid: ", pid)
			return
		}
	}

	// start server
	if !daemon {
		setPid(os.Getpid())
		log.RegisterExitHandler(removePid)

		os.Remove(errFile())

		defer func() {
			if err := recover(); err != nil {
				stdLog.Printf("panic recovered, err: %s", err)
			}
			removePid()
		}()

		if err := server.Start(); err != nil {
			os.WriteFile(errFile(), []byte(err.Error()), 0666)
			stdLog.Print(err)
		}
		return
	}

	// start server as a daemon
	cmd := &exec.Cmd{
		Path: os.Args[0],
		Args: []string{os.Args[0], "start"},
		Env:  os.Environ(),
	}

	stdWriter, err := os.OpenFile("std.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		stdLog.Print("open std.log err:", err)
		return
	}

	cmd.Stderr = stdWriter
	cmd.Stdout = stdWriter
	err = cmd.Start()
	if err != nil {
		stdLog.Print("start process err:", err)
		return
	}

	stdLog.Printf("server started in the background with pid %d", cmd.Process.Pid)
	stdLog.Print("std logs will be redirect to local file std.log")

	for i := 0; i < 3; i++ {
		stdLog.Printf("check health in %ds", 3-i)
		time.Sleep(time.Second)
		b, _ := os.ReadFile(errFile())
		if len(b) > 0 {
			stdLog.Fatalf("%s", b)
		}
	}

	stdLog.Print("server status is OK")
}

func errFile() string {
	return "/tmp/keepshare.err"
}

func pidFile() string {
	return "/tmp/keepshare.pid"
}

func getPid() int {
	file := pidFile()
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return -1
	}
	if err != nil {
		stdLog.Fatal("stat pid file err:", err)
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		stdLog.Fatal("read pid file err:", err)
	}
	pid, err := strconv.Atoi(string(bytes))
	if err != nil {
		stdLog.Fatal("parse pid data err:", err)
	}
	return pid
}

func setPid(pid int) {
	file := pidFile()
	err := os.WriteFile(file, []byte(strconv.Itoa(pid)), 0666)
	if err != nil {
		stdLog.Print("failed to record pid, you may not be able to stop the program with `./keepshare stop`")
	}
}

func removePid() {
	os.Remove(pidFile())
}
