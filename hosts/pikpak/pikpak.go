// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"embed"
	"fmt"

	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/account"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/hibiken/asynq"
)

// PikPak official website: https://mypikpak.com/
type PikPak struct {
	*hosts.Dependencies

	q   *query.Query
	m   *account.Manager
	api *api.API
}

//go:embed  rawsql/*.sql
var sqlFS embed.FS

func init() {
	sql, err := hosts.ReadSQLFileFromFS(sqlFS)
	if err != nil {
		panic(fmt.Errorf("read sql files err: %w", err))
	}

	hosts.Register(&hosts.Properties{Name: "pikpak", New: New, CreateTableStatements: sql})
}

// New create a PikPak host.
func New(d *hosts.Dependencies) hosts.Host {
	p := &PikPak{Dependencies: d}

	p.q = query.Use(p.Mysql)

	p.api = api.New(p.q, d)

	p.m = account.NewManager(p.q, p.api, d)

	go p.deleteFilesBackground()

	d.Queue.RegisterHandler(taskTypeSyncWorkerInfo, asynq.HandlerFunc(p.syncWorkerHandler))

	return p
}
