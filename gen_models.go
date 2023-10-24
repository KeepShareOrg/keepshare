// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build ignore

// Run it to generate golang models into dir gorm-gen/model/ and gorm-gen/query/
// Example:
//   go run gen_models.go --mysql="root:admin@(127.0.0.1:3306)/keepshare" --prefix=keepshare --out-path=server/query
//   go run gen_models.go --mysql="root:admin@(127.0.0.1:3306)/keepshare" --prefix=pikpak --out-path=hosts/pikpak/query

package main

import (
	"flag"
	"log"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	var dsn, prefix, path string
	flag.StringVar(&dsn, "mysql", "root:admin@(127.0.0.1:3306)/keepshare", "mysql DNS, database and tables must have been created")
	flag.StringVar(&prefix, "prefix", "", "filter tables name with this prefix")
	flag.StringVar(&path, "out-path", "", "the path where the generated files are stored")
	flag.Parse()

	log.SetFlags(0)

	if dsn == "" || prefix == "" || path == "" {
		log.Println("All flags are required!")
		log.Println()
		flag.PrintDefaults()
		log.Print(`
Example:
  go run gen_models.go --mysql="root:admin@(127.0.0.1:3306)/keepshare" --prefix=keepshare --out-path=server/query
  go run gen_models.go --mysql="root:admin@(127.0.0.1:3306)/keepshare" --prefix=pikpak --out-path=hosts/pikpak/query
`)
		return
	}

	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		log.Fatal("open mysql err:", err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath: path,
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db)

	tables, err := db.Migrator().GetTables()
	if err != nil {
		log.Fatal("get all tables fail:", err)
	}

	for _, tableName := range tables {
		if strings.HasPrefix(strings.ToLower(tableName), strings.ToLower(prefix)) {
			//g.ApplyBasic(g.GenerateModel(tableName))
			modelName := db.Config.NamingStrategy.SchemaName(tableName)
			modelName = trimPrefix(modelName, prefix)
			log.Print("modelName:", modelName)
			g.ApplyBasic(g.GenerateModelAs(tableName, modelName))
		}
	}

	// Generate the code
	g.Execute()
}

func trimPrefix(s, prefix string) string {
	if prefix == "" || len(s) < len(prefix) {
		return s
	}
	if strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):]
	}
	return s
}
