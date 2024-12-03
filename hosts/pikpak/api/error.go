// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"strings"
)

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

// IsHasJoinedReferralErr return whether the error is user has joined the referral program
func IsHasJoinedReferralErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "has_joined_referral")
}

var (
	EmptyPasswordErr = errors.New("password is empty")
)

// IsEmptyPasswordErr returns whether the error is caused by empty password.
func IsEmptyPasswordErr(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, EmptyPasswordErr)
}

// IsInvalidGrantErr returns whether the error is caused by invalid grant.
func IsInvalidGrantErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "invalid_grant")
}

// IsSubTaskNumberLimitErr returns whether the error is caused by sub task number limit.
func IsSubTaskNumberLimitErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "task_sub_task_num_limit")
}

// IsFileNameEmptyErr returns whether the error is caused by file name empty.
func IsFileNameEmptyErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "file_name_empty")
}

// IsTaskUrlResolveError returns whether the error is caused by task url resolve error.
func IsTaskUrlResolveError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "task_url_resolve_error")
}

// IsSpaceFolderNotExistError returns whether the error is caused by space folder not exist.
func IsSpaceFolderNotExistError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "space_folder_not_exist")
}

// IsFileNotFoundError returns whether the error is caused by file not found.
func IsFileNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "file_not_found")
}

// IsInternalError returns whether the error is caused by internal error.
func IsInternalError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "internal")
}

// IsCaptchaInvalidError returns whether the error is caused by captcha invalid.
func IsCaptchaInvalidError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "captcha_invalid")
}

// IsShouldNotRetryError returns whether the error is should not retry.
func IsShouldNotRetryError(err error) bool {
	if err == nil {
		return false
	}

	return IsTaskUrlResolveError(err) || IsFileNameEmptyErr(err) || IsSubTaskNumberLimitErr(err)
}
