// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package comm

// enum all statuses.
const (
	StatusOK      = "PHASE_TYPE_COMPLETE"
	StatusError   = "PHASE_TYPE_ERROR"
	StatusRunning = "PHASE_TYPE_RUNNING"
	StatusPending = "PHASE_TYPE_PENDING"
)

// enum PikPak link status.
const (
	LinkStatusOK      = "OK"
	LinkStatusUnknown = "UNKNOWN"
	LinkStatusLimited = "LIMITED"
	LinkStatusError   = "ERROR"
)

// IsFinalStatus returns whether this status is a final state which can not be updated anymore.
func IsFinalStatus(s string) bool {
	return s == StatusOK || s == StatusError
}

const (
	// MaxPremiumWorkers is the maximum number of premium workers bound to a master.
	MaxPremiumWorkers = 50

	// SlowTaskTriggerConditionTimes is the maximum number of times to trigger a slow task.
	SlowTaskTriggerConditionTimes = 2
)

const (
	// RedeemCodeStatusNotUsed is the status of a redeem code which is not used.
	RedeemCodeStatusNotUsed = "NOT_USED"
	// RedeemCodeStatusUsed is the status of a redeem code which is used.
	RedeemCodeStatusUsed = "USED"
	// RedeemCodeInvalid is the status of a redeem code which is invalid.
	RedeemCodeInvalid = "INVALID"
)
