// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import "testing"

func TestNewID(t *testing.T) {
	for i := 0; i < 10; i++ {
		uid := NewID()
		t.Log(uid, len(uid))
	}
}

func TestNewChannelID(t *testing.T) {
	for i := 0; i < 10; i++ {
		uid := NewChannelId()
		t.Log(uid, len(uid))
	}
}
