// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package async

import (
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	fChan            chan func()
	fChanInitOnce    sync.Once
	fChanConcurrency = runtime.NumCPU()
	fChanSize        = fChanConcurrency * 1024
)

// Run functions background with a fixed number of goroutines.
func Run(fs ...func()) {
	fChanInitOnce.Do(func() {
		fChan = make(chan func(), fChanSize)
		for i := 0; i < fChanConcurrency; i++ {
			go func() {
				for f := range fChan {
					runWithTimeout(f, 10*time.Second)
				}
			}()
		}
	})

	for _, f := range fs {
		select {
		case fChan <- f:
		default:
			log.Warnf("function chan maybe full, size: %d, concurrency: %d", fChanSize, fChanConcurrency)
			return
		}
	}
}

func runWithTimeout(f func(), timeout time.Duration) {
	done := make(chan struct{})
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	go func() {
		f() // f can not be canceled
		done <- struct{}{}
	}()

	select {
	case <-done:
		return
	case <-timer.C:
		log.Warnf("function timeout in %s", timeout)
		return
	}
}
