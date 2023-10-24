// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package hosts

import (
	"context"
	"errors"
	"io/fs"
	"regexp"
	"strings"

	"github.com/KeepShareOrg/keepshare/pkg/mail"
	"github.com/KeepShareOrg/keepshare/pkg/queue"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Host is an interface of file hosting provider.
type Host interface {
	// CreateFromLinks create shared links based on the input original links.
	CreateFromLinks(ctx context.Context, userID string, originalLinks []string, createBy string) (sharedLinks map[string]*share.Share, err error)

	// GetStatuses return the statuses of each host shared link.
	GetStatuses(ctx context.Context, userID string, hostSharedLinks []string) (statuses map[string]share.State, err error)

	// GetStatistics return the statistics of each host shared link.
	GetStatistics(ctx context.Context, userID string, hostSharedLinks []string) (details map[string]share.Statistics, err error)

	// Delete delete shared links by original links.
	Delete(ctx context.Context, userID string, originalLinks []string) error

	// HostInfo returns basic information of the host.
	HostInfo(ctx context.Context, userID string, options map[string]any) (resp map[string]any, err error)
}

// Properties of a host.
type Properties struct {
	// Name is the host's name.
	// The name must match this regular expression: `^[0-9A-Za-z_]{1,32}$`
	Name string

	// New is the function to create a Host instance.
	New func(d *Dependencies) Host

	// CreateTableStatements return the statements to create tables in mysql database.
	// The name of tables must be prefixed with the host provider's name.
	// If the host do not need mysql database, this property can be empty.
	CreateTableStatements []string
}

// Dependencies of hosts.
type Dependencies struct {
	Mysql  *gorm.DB
	Redis  *redis.Client
	Mailer mail.Mailer
	Queue  *queue.Client
}

var hosts = map[string]*HostWithProperties{}

// HostWithProperties save properties for hosts.
type HostWithProperties struct {
	Host
	p *Properties
}

const namePattern = `^[0-9A-Za-z_]{1,32}$`

// Errors that may occur when registering a host provider.
var (
	ErrNameAlreadyRegistered = errors.New("name already registered")
	ErrInvalidName           = errors.New("name does not match this regular expression: " + namePattern)
)

// Register a new host.
func Register(p *Properties) error {
	if match, _ := regexp.MatchString(namePattern, p.Name); !match {
		return ErrInvalidName
	}

	name := strings.ToLower(p.Name)
	if _, ok := hosts[name]; ok {
		return ErrNameAlreadyRegistered
	}

	hosts[name] = &HostWithProperties{p: p}
	return nil
}

// Start all hosts.
func Start(d *Dependencies) {
	for _, v := range hosts {
		if v.Host == nil {
			v.Host = v.p.New(d)
		}
	}
}

// Get a host provider by name.
func Get(name string) *HostWithProperties {
	return hosts[strings.ToLower(name)]
}

// GetAll return all host providers.
func GetAll() []*HostWithProperties {
	var all []*HostWithProperties
	for _, v := range hosts {
		all = append(all, v)
	}
	return all
}

// Name returns the host's name.
func (h *HostWithProperties) Name() string {
	return h.p.Name
}

// CreateTableStatements returns the statements to create tables of the host.
func (h *HostWithProperties) CreateTableStatements() []string {
	return h.p.CreateTableStatements
}

// ReadSQLFileFromFS walk the directory and read *.sql files from FS such as embed.FS.
func ReadSQLFileFromFS(fsys fs.FS) ([]string, error) {
	var data []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		// filter dir and other files
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// read .sql file
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		data = append(data, string(b))
		return nil
	})
	return data, err
}
