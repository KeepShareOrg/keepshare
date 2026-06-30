// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package forward

import (
	"strings"

	"github.com/samber/lo"
)

var PikPakEmailList = []string{
	"@accounts.mypikpak.com",
	"@mypikpak.net",
}

// Filter reports whether msg should be forwarded. masters is the set of
// currently-registered pikpak master emails; it is refreshed by Run.
//
// On accept it returns the matched master-account email (the value stored in
// pikpak_master_account), which the caller uses to resolve the keepshare user
// and substitute the notice's <TO> token. A recipient is matched either by
// exact equality or by local-part against the masters set.
func Filter(msg Message, masters map[string]struct{}) (matched string, reason string, accept bool) {
	valid := lo.SomeBy(PikPakEmailList, func(item string) bool {
		return strings.Contains(msg.From, item)
	})
	if !valid {
		return "", "from not noreply@accounts.mypikpak.com", false
	}

	if msg.Mailbox == "" {
		return "", "mailbox not found", false
	}

	for m := range masters {
		if i := strings.LastIndex(m, "@"); i > 0 {
			localParts := m[:i]
			if localParts == msg.Mailbox {
				return m, "", true
			}
		}
	}

	return "", "recipient not in pikpak_master_account", false
}
