// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/server/rawsql"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	var dropEmpty bool

	cmd := &cobra.Command{
		Use:   "tables",
		Short: "Operations related to database and tables.\nThe mysql address must be specified in the configuration file or environment variable, see more: `keepshare config`",
	}

	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "create tables",
		Example: "keepshare tables --create table1 table2 ...",
		Run: func(cmd *cobra.Command, args []string) {
			createTables(cmd, args, dropEmpty)
		},
	}

	createCmd.Flags().BoolVar(&dropEmpty, "drop-empty", false, "When creating a table, if the table is empty, drop the table and recreate it.")

	dumpCmd := &cobra.Command{
		Use:     "dump",
		Short:   "Dump the CREATE TABLE statements for tables",
		Example: "keepshare tables dump table1 table2 ...",
		Run:     dumpTables,
	}

	cmd.AddCommand(createCmd, dumpCmd)
	rootCmd.AddCommand(cmd)
}

func createTables(_ *cobra.Command, args []string, dropEmpty bool) {
	if err := config.Load(); err != nil {
		stdLog.Fatal("load config err:", err)
	}

	db := config.MySQL()
	var current []string
	if err := db.Raw("SHOW TABLES").Scan(&current).Error; err != nil {
		stdLog.Fatal("get current tables err:", err)
	}

	for t, s := range loadStatements() {
		if len(args) > 0 && !lo.Contains(args, t) {
			continue
		}

		if lo.Contains(current, t) {
			if !dropEmpty {
				stdLog.Printf("IGNORE exists table `%s`", t)
				continue
			}

			// check if table is empty
			var n int
			err := db.Raw(fmt.Sprintf("SELECT 1 FROM `%s` LIMIT 1", t)).Scan(&n).Error
			if err != nil {
				stdLog.Fatal(err)
			}
			// not empty
			if n == 1 {
				stdLog.Printf("IGNORE exists and NON-EMPTY table `%s`", t)
				continue
			}
			//is empty, drop it
			err = db.Exec(fmt.Sprintf("DROP TABLE `%s`", t)).Error
			if err != nil {
				stdLog.Fatalf("DROP TABLE `%s` err: %v", t, err)
			}
			stdLog.Printf("DROP TABLE `%s`", t)
		}

		err := db.Exec(s).Error
		if err != nil {
			stdLog.Fatalf("CREATE TABLE `%s` err: %v", t, err)
		} else {
			stdLog.Printf("CREATE TABLE `%s`", t)
		}
	}
}

func dumpTables(_ *cobra.Command, args []string) {
	for t, s := range loadStatements() {
		if len(args) > 0 && !lo.Contains(args, t) {
			continue
		}

		stdLog.Println(s)
		stdLog.Println()
	}
}

func loadStatements() map[string]string {
	all := map[string]string{}

	// server tables
	stmts, err := hosts.ReadSQLFileFromFS(rawsql.FS)
	if err != nil {
		stdLog.Fatal("read rawsql.FS err:", err)
	}
	for table, stmt := range splitTables(stmts...) {
		if !strings.HasPrefix(strings.ToLower(table), "keepshare_") {
			stdLog.Fatalf("invalid table name `%s` from rawsql.FS, the name of tables must be prefixed with the 'keepshare_'.", table)
		}
		all[table] = stmt
	}

	// host providers tables
	for _, host := range hosts.GetAll() {
		stmts := host.CreateTableStatements()
		if len(stmts) == 0 {
			continue
		}
		for table, stmt := range splitTables(stmts...) {
			if !strings.HasPrefix(strings.ToLower(table), strings.ToLower(host.Name())) {
				stdLog.Fatalf("invalid table name `%s` from host `%s`, the name of tables must be prefixed with the host provider's name.", table, host.Name())
			}
			all[table] = stmt
		}
	}

	return all
}

func splitTables(stmts ...string) map[string]string {
	statements := strings.Join(stmts, "\n")
	m := map[string]string{}

	// replace spaces
	replace := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+`)
	statements = replace.ReplaceAllString(statements, "CREATE TABLE ")

	// split
	r1 := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(IF\s+NOT\s+EXISTS\s+)?` + "[`_A-Za-z0-9]+")
	r2 := regexp.MustCompile("[`_A-Za-z0-9]+$")

	for _, s := range strings.Split(statements, "CREATE TABLE ") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		s = "CREATE TABLE " + s
		table := r2.FindString(r1.FindString(s))
		table = strings.TrimPrefix(table, "`")
		table = strings.TrimSuffix(table, "`")
		m[table] = s
	}

	return m
}
