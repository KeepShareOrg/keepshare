// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const mod = 100_000

var (
	sequence = rand.New(rand.NewSource(time.Now().UnixNano())).Int63() % mod
	lock     = &sync.Mutex{}
)

// NewID generate new user id.
func NewID() string {
	lock.Lock()
	defer lock.Unlock()

	// make sure that no duplicate id will be generated within 1 millisecond.
	time.Sleep(time.Millisecond * 3 / mod)

	sequence++
	n := time.Now().UnixMilli()*mod + sequence%mod
	return strconv.FormatInt(n, 32)
}

// NewChannelId generate new channel id.
func NewChannelId() string {
	id := NewID()
	if len(id) >= 8 {
		return id[len(id)-8:]
	}
	return ""
}
