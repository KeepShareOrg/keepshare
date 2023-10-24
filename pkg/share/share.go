// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package share

import (
	"time"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
)

// State is the current status of some shared link.
type State string

// Enum all Statuses.
const (
	StatusPending   State = "PENDING" // batch add "link to share" will create a pending task first
	StatusUnknown   State = "UNKNOWN"
	StatusCreated   State = "CREATED"
	StatusOK        State = "OK"
	StatusDeleted   State = "DELETED"
	StatusNotFound  State = "NOT_FOUND"
	StatusSensitive State = "SENSITIVE" // the content corresponding to the link contains sensitive resources.
	StatusBlocked   State = "BLOCKED"   // blocked by user self.
	StatusError     State = "ERROR"
)

const (
	// AutoShare indicates that the sharing link was created automatically.
	AutoShare = "Auto Share"
	// LinkToShare indicates that the sharing link was created manually.
	LinkToShare = "Link to Share"
)

// Share contains all information of a shared link.
type Share struct {
	// State is the current status of the shared link.
	State State
	// Title is the name or title of the shared link.
	Title string
	// HostSharedLink is the shared link that can be used to share and forward to others.
	HostSharedLink string
	// OriginalLink is the original link used to generate this share.
	OriginalLink string
	// CreatedBy is the creator of this shared link.
	CreatedBy string
	// CreatedAt is the time when the shared link was created
	CreatedAt time.Time
	// Size is the total size of all files included in this shared link.
	Size int64
	// Statistics is the statistical data of this shared link
	Statistics
}

// Statistics is the statistical data of this shared link.
type Statistics struct {
	// Visitor is the number of times the shared link was viewed by other users
	Visitor int32
	// Stored is the number of times the shared link was saved by other users
	Stored int32
	// Revenue is the total revenue of this shared link in SGD cents.
	Revenue int64
}

func (s State) String() string {
	return string(s)
}

// StatusFromFileStatus returns the status of shared link corresponding to the status of file.
func StatusFromFileStatus(s string) State {
	switch s {
	case comm.StatusOK:
		return StatusOK
	case comm.StatusError:
		return StatusSensitive
	case comm.StatusRunning:
		return StatusCreated
	default:
		return StatusSensitive
	}
}
