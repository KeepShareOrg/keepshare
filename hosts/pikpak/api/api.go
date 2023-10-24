// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package api

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/coocood/freecache"
	"github.com/go-resty/resty/v2"
)

// configs.
const (
	apiServer      = "https://api-drive.mypikpak.com"
	userServer     = "https://user.mypikpak.com"
	clientID       = "YNxT9w7GMdWvEOKa"
	deviceID       = "c858a46bfca5c5f61b1702ed6c303acb"
	acceptLanguage = "en,en-US;q=0.9"
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"

	webClientID = "YUMx5nI8ZU8Ap8pm"
)

// API PikPak server api.
type API struct {
	q     *query.Query
	cache *freecache.Cache

	*hosts.Dependencies
}

// New returns server api instance.
func New(q *query.Query, d *hosts.Dependencies) *API {
	api := &API{
		q:            q,
		Dependencies: d,
		cache:        freecache.NewCache(50 * 1024 * 1024),
	}

	go api.checkFilesBackground()
	go api.updatePremiumExpirationBackground()

	return api
}

type (
	// JSON alias of string map.
	JSON map[string]any

	// RespErr server response error struct.
	RespErr struct {
		ErrorKey         string `json:"error"`
		ErrorCode        int    `json:"error_code"`
		ErrorDescription string `json:"error_description"`
	}
)

var resCli = resty.New().
	SetHeader("X-Client-Id", clientID).
	SetHeader("X-Device-Id", deviceID).
	SetHeader("Accept-Language", acceptLanguage).
	SetHeader("User-Agent", userAgent).
	SetHeader("Content-Type", "application/json").
	SetTimeout(time.Second * 10).
	SetRetryCount(1)

func apiURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}
	return fmt.Sprintf("%s/%s", apiServer, strings.TrimLeft(path, "/"))
}

func userURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}
	return fmt.Sprintf("%s/%s", userServer, strings.TrimLeft(path, "/"))
}

// Error check error and implement error interface..
func (e *RespErr) Error() error {
	if e == nil || e.ErrorKey == "" || strings.EqualFold(e.ErrorKey, "OK") {
		return nil
	}
	if (e.ErrorKey != "" && !strings.EqualFold(e.ErrorKey, "OK")) || e.ErrorCode > 0 {
		if e.ErrorDescription != "" {
			return fmt.Errorf("%d|%s|%s", e.ErrorCode, e.ErrorKey, e.ErrorDescription)
		}
		return fmt.Errorf("%d|%s", e.ErrorCode, e.ErrorKey)
	}
	return nil
}

var (
	apiRand       = rand.New(rand.NewSource(time.Now().UnixNano()))
	emailSequence = apiRand.Uint64()
)

func randomPassword() string {
	return fmt.Sprintf("#Pk.%d", apiRand.Uint32())
}

func randomDevice() string {
	s := fmt.Sprintf("Device.%d.%d", apiRand.Uint64(), time.Now().UnixNano())
	r := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", r)
}

func (api *API) randomEmail() string {
	const mod = 1000000
	seq := atomic.AddUint64(&emailSequence, 1)
	n := uint64(time.Now().UnixMilli())*mod + seq%mod
	return strconv.FormatUint(n, 32) + "@" + api.Mailer.Domain()
}
