// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "keepshare",
	Short: "File hosting and sharing automation",
}

var stdLog = log.New(os.Stdout, "", 0)

// Execute root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		stdLog.Fatal(err)
	}
}
