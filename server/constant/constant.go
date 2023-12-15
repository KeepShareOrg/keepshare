// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package constant

// constant keys.
var (
	UserID      = "user_id"
	DeviceID    = "device_id"
	RequestID   = "request_id"
	Channel     = "channel"
	IP          = "ip"
	Error       = "error"
	Message     = "message"
	SharedLink  = "shared_link"
	ShareStatus = "status"
	Email       = "email"
	Username    = "username"
	Link        = "link"
	Host        = "host"

	HeaderDeviceID = "X-Device-Id"
)

// about email verification
var (
	EmailVerificationDone       int32 = 1
	EmailVerificationUnComplete int32 = 0
)

type VerificationAction string

const (
	VerificationActionResetPassword  VerificationAction = "reset_password"
	VerificationActionChangeEmail    VerificationAction = "change_email"
	VerificationActionChangePassword VerificationAction = "change_password"
)
