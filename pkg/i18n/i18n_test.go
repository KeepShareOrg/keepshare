// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package i18n

import (
	"context"
	"testing"

	"github.com/KeepShareOrg/keepshare/locale"
)

func TestGet(t *testing.T) {
	err := Load(locale.FS)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("loaded:", Languages())

	msg, err := Get(context.Background(), "internal", WithLanguages("en"), WithDataMap("error", "test"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msg)
}
