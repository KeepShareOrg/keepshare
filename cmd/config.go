// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show available configurations",
		Long:  config.Help(),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.Help())
		},
	}

	rootCmd.AddCommand(cmd)
}
