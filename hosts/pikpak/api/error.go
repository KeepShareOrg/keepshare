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

var accountLimitedErrors = []string{
	"task_daily_create_limit",
	"task_run_nums_limit",
}

// IsAccountLimited returns whether the error is caused by space not enough.
func IsAccountLimited(err error) bool {
	if err == nil {
		return false
	}

	if IsSpaceNotEnoughErr(err) {
		return true
	}

	msg := err.Error()
	for _, v := range accountLimitedErrors {
		if strings.Contains(msg, v) {
			return true
		}
	}

	return false
}
