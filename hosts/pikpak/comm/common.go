// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package comm

import "time"

// enum all statuses.
const (
	StatusOK      = "PHASE_TYPE_COMPLETE"
	StatusError   = "PHASE_TYPE_ERROR"
	StatusRunning = "PHASE_TYPE_RUNNING"
)

// IsFinalStatus returns whether this status is a final state which can not be updated anymore.
func IsFinalStatus(s string) bool {
	return s == StatusOK || s == StatusError
}

const (
	// CreateFileSyncWaitSeconds is the maximum wait time in seconds when creating a file from link.
	CreateFileSyncWaitSeconds = 2

	// RunningFilesMaxAge is the maximum time for a file from running to complete.
	// After this time, set the file to error status.
	RunningFilesMaxAge = 48 * time.Hour

	// RunningFilesSelectLimit is the maximum number of running files selected from mysql each time.
	RunningFilesSelectLimit = 100

	// MaxPremiumWorkers is the maximum number of premium workers bound to a master.
	MaxPremiumWorkers = 10
)
