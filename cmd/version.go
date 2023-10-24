// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version app version, set by go build -ldflags=
	Version = "SET_ON_BUILD"

	// Commit git commit.
	Commit = "SET_ON_BUILD"

	// Build environments for this app, set by go build -ldflags=
	Build = "SET_ON_BUILD"
)

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show app version",
		Long:  "show app version",
		Run: func(cmd *cobra.Command, _ []string) {
			if Version == "" {
				Version = "dev-" + Commit
			}
			if short, _ := cmd.Flags().GetBool("short"); short {
				fmt.Println(Version)
			} else {
				fmt.Printf("version: %s\ncommit: %s\nbuild: %s\n", Version, Commit, Build)
			}
		},
	}

	cmd.Flags().BoolP("short", "s", false, "Only print the version number, ignoring other information")

	rootCmd.AddCommand(cmd)
}
