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

// IsTaskRunNumsLimitErr returns whether the error is caused by task run nums limit.
func IsTaskRunNumsLimitErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "task_run_nums_limit")
}

// IsTaskDailyCreateLimitErr returns whether the error is caused by task daily create limit.
func IsTaskDailyCreateLimitErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "task_daily_create_limit")
}

// IsAccountLimited returns whether the error is caused by space not enough or rate limit...
func IsAccountLimited(err error) bool {
	if err == nil {
		return false
	}

	if IsSpaceNotEnoughErr(err) || IsTaskDailyCreateLimitErr(err) || IsTaskRunNumsLimitErr(err) {
		return true
	}

	return false
}

// IsInvalidAccountOrPasswordErr returns whether the error is caused by invalid account or password.
func IsInvalidAccountOrPasswordErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "invalid_account_or_password")
}

// IsInvalidRedeemCodeErr returns whether the error is caused by invalid redeem code.
func IsInvalidRedeemCodeErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "invalid_activation_code")
}

// IsMemberCodeNotSupportRegionErr returns whether the error is caused by member code not support region.
func IsMemberCodeNotSupportRegionErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "member_code_not_support_region")
}

// IsSlowTaskLinkError returns whether the error is caused by slow task link.
func IsSlowTaskLinkError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "slow_task_link")
}
