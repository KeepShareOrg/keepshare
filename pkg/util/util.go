// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"strconv"
)

// enum file size.
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	_  = TB
)

// ToJSON marshal data to JSON, ignore errors.
func ToJSON(v any, pretty ...bool) string {
	var bs []byte
	if len(pretty) > 0 && pretty[0] {
		bs, _ = json.MarshalIndent(v, "", "  ")
	} else {
		bs, _ = json.Marshal(v)
	}
	return string(bs)
}

// Atoi parse string to int, ignore errors.
func Atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// FirstNotEmpty returns the first value not equals to the zero value.
func FirstNotEmpty[T comparable](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}
