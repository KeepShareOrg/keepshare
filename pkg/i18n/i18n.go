// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package i18n

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var bundle = i18n.NewBundle(language.English)

// Load all message files in dir.
func Load(fs fs.ReadDirFS) error {
	dir, err := fs.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read dir err: %w", err)
	}

	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	for _, entry := range dir {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}

		if _, err := bundle.LoadMessageFileFS(fs, name); err != nil {
			return fmt.Errorf("load message file %s err: %w", name, err)
		}
	}

	return nil
}

// Get localized message by languages.
func Get(ctx context.Context, id string, opts ...Option) (string, error) {
	c := &options{
		languages:      nil,
		LocalizeConfig: &i18n.LocalizeConfig{MessageID: id},
	}
	for _, o := range opts {
		o(c)
	}
	if accept := ctx.Value(acceptLanguage(0)); accept != nil {
		c.languages = append(c.languages, accept.(string))
	}

	return i18n.NewLocalizer(bundle, c.languages...).Localize(c.LocalizeConfig)
}

// MustGet get localized message by languages, ignore error.
// If an error occurs, returns the id.
func MustGet(ctx context.Context, id string, options ...Option) string {
	s, err := Get(ctx, id, options...)
	if err != nil {
		return id
	}
	return s
}

type acceptLanguage int

// ContextWithAcceptLanguage store accept language into context.
func ContextWithAcceptLanguage(ctx context.Context, accept string) context.Context {
	return context.WithValue(ctx, acceptLanguage(0), accept)
}

// Option to get localized messages.
type Option func(o *options)

type options struct {
	languages []string
	*i18n.LocalizeConfig
}

// WithLanguages set plural count.
func WithLanguages(languages ...string) Option {
	return func(o *options) {
		o.languages = append(o.languages, languages...)
	}
}

// WithCount set plural count.
func WithCount(count any) Option {
	return func(o *options) {
		o.PluralCount = count
	}
}

// WithData set template data.
func WithData(data any) Option {
	return func(o *options) {
		o.TemplateData = data
	}
}

// WithDataMap set template data.
func WithDataMap(kvs ...string) Option {
	if len(kvs)%2 != 0 {
		panic(errors.New("kvs must be passed in pairs"))
	}
	m := make(map[string]string, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		m[kvs[i]] = kvs[i+1]
	}
	return func(o *options) {
		o.TemplateData = m
	}
}

// Languages returns all languages loaded.
func Languages() []string {
	var a []string
	for _, v := range bundle.LanguageTags() {
		a = append(a, v.String())
	}
	return a
}
