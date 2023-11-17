// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"time"
)

func init() {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart server",
		Run: func(_ *cobra.Command, _ []string) {
			stop()
			time.Sleep(time.Millisecond * 500)
			start(true)
		},
	}
	rootCmd.AddCommand(cmd)
}
