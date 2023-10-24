// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package static

import "embed"

// FS holds all locale files.
//
//go:embed dist
var FS embed.FS
