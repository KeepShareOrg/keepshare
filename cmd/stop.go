// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop server",
		Run: func(_ *cobra.Command, _ []string) {
			stop()
		},
	}

	rootCmd.AddCommand(cmd)
}

func stop() {
	// check if the server is running
	pid := getPid()
	if pid == -1 {
		stdLog.Println("server is not running")
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		stdLog.Println("server is not running", err)
		return
	}
	err = process.Kill()
	switch err {
	case os.ErrProcessDone:
		stdLog.Printf("process with pid %d already finished", pid)
	case nil:
		stdLog.Printf("process with pid %d has been killed", pid)
	default:
		stdLog.Printf("kill process with pid:%d, err:%v", pid, err)
	}

	file := pidFile()
	if err = os.Remove(file); err != nil {
		stdLog.Printf("remove pid file:%s, err:%v", file, err)
	}
}
