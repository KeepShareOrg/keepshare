// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import "strings"

// IsSpaceNotEnoughErr returns whether the error is caused by space not enough.
func IsSpaceNotEnoughErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "file_space_not_enough")
}

// IsAccountLimited returns whether the error is caused by space not enough.
func IsAccountLimited(err error) bool {
	if err == nil {
		return false
	}

	if IsSpaceNotEnoughErr(err) {
		return true
	}

	if strings.Contains(err.Error(), "task_daily_create_limit") {
		return true
	}

	// TODO other errors

	return false
}
