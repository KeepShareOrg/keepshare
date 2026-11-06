// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"regexp"
	"sort"
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

	statements := loadStatements()
	tables := sortedTableNames(statements, args)
	existing := make(map[string]bool, len(current))
	for _, t := range current {
		existing[t] = true
	}

	if dropEmpty {
		// Drop children before parents, then recreate in the opposite order.
		for i := len(tables) - 1; i >= 0; i-- {
			t := tables[i]
			if !existing[t] {
				continue
			}
			var n int
			if err := db.Raw(fmt.Sprintf("SELECT 1 FROM `%s` LIMIT 1", t)).Scan(&n).Error; err != nil {
				stdLog.Fatal(err)
			}
			if n != 0 {
				continue
			}
			if t == "keepshare_shared_link" && existing["keepshare_shared_link_visitor_ip"] {
				stdLog.Printf("IGNORE table `%s` referenced by keepshare_shared_link_visitor_ip; select both tables to recreate", t)
				continue
			}
			if err := db.Exec(fmt.Sprintf("DROP TABLE `%s`", t)).Error; err != nil {
				stdLog.Fatalf("DROP TABLE `%s` err: %v", t, err)
			}
			delete(existing, t)
			stdLog.Printf("DROP TABLE `%s`", t)
		}
	}

	for _, t := range tables {
		if existing[t] {
			stdLog.Printf("IGNORE exists table `%s`", t)
			continue
		}
		if err := db.Exec(statements[t]).Error; err != nil {
			stdLog.Fatalf("CREATE TABLE `%s` err: %v", t, err)
		}
		stdLog.Printf("CREATE TABLE `%s`", t)
	}
}

func dumpTables(_ *cobra.Command, args []string) {
	statements := loadStatements()
	for _, t := range sortedTableNames(statements, args) {
		stdLog.Println(statements[t])
		stdLog.Println()
	}
}

func sortedTableNames(statements map[string]string, args []string) []string {
	tables := make([]string, 0, len(statements))
	for t := range statements {
		if len(args) == 0 || lo.Contains(args, t) {
			tables = append(tables, t)
		}
	}
	// keepshare_shared_link must precede its foreign-key child tables.
	sort.Strings(tables)
	return tables
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
