// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package forward

import (
	"sync"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
)

// Rotator picks the next SMTP sender account in a round-robin fashion,
// skipping any account that recently failed and is still cooling down.
type Rotator struct {
	mu       sync.Mutex
	accounts []config.ForwardAccount
	cooldown time.Duration
	next     int
	until    map[string]time.Time
}

// NewRotator constructs a Rotator over the supplied accounts.
func NewRotator(accounts []config.ForwardAccount, cooldown time.Duration) *Rotator {
	return &Rotator{
		accounts: accounts,
		cooldown: cooldown,
		until:    make(map[string]time.Time),
	}
}

// Next returns the next usable account and a flag indicating whether one
// was available. When all accounts are cooling down, ok is false and the
// returned account is the zero value.
func (r *Rotator) Next() (config.ForwardAccount, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	n := len(r.accounts)
	for i := 0; i < n; i++ {
		idx := (r.next + i) % n
		acc := r.accounts[idx]
		if until, ok := r.until[acc.Email]; ok && until.After(now) {
			continue
		}
		r.next = (idx + 1) % n
		return acc, true
	}
	return config.ForwardAccount{}, false
}

// MarkFailed puts the account into cooldown.
func (r *Rotator) MarkFailed(email string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.until[email] = time.Now().Add(r.cooldown)
}

// MarkOK clears any recorded cooldown for the account.
func (r *Rotator) MarkOK(email string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.until, email)
}
