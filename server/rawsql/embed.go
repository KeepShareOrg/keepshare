// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rawsql

import "embed"

// FS holds all sql files.
//
//go:embed *.sql
var FS embed.FS
