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

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/samber/lo"
	"gorm.io/gen"
	"gorm.io/gorm/clause"

	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/coocood/freecache"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

// configs.
const (
	referralServer = "https://api-referral.mypikpak.com"
	clientID       = "YNxT9w7GMdWvEOKa"
	acceptLanguage = "en,en-US;q=0.9"

	webClientID = "YUMx5nI8ZU8Ap8pm"
)

var (
	// userServer and apiServer can be overridden by `pikpak.user_server` and
	// `pikpak.api_server` in the configuration file.
	userServer = "https://user.mypikpak.com"
	apiServer  = "https://api-drive.mypikpak.com"

	deviceID  = "c858a46bfca5c5f61b1702ed6c303acb"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"

	resCli = resty.New().
		SetHeader("X-Device-Id", deviceID).
		SetHeader("User-Agent", userAgent).
		SetHeader("X-Client-Id", clientID).
		SetHeader("Accept-Language", acceptLanguage).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Second * 10).
		SetRetryCount(1)
)

type createLinkLimitData struct {
	IpNum    uint32 `json:"ip_num"`
	UnitTime uint32 `json:"unit_time"`
}

// API PikPak server api.
type API struct {
	q     *query.Query
	cache *freecache.Cache

	*hosts.Dependencies
	createLinkLimitOptions map[string]createLinkLimitData
}

// New returns server api instance.
func New(q *query.Query, d *hosts.Dependencies) *API {
	api := &API{
		q:            q,
		Dependencies: d,
		cache:        freecache.NewCache(50 * 1024 * 1024),
	}

	if v := viper.GetString("pikpak.user_server"); v != "" {
		userServer = v
	}
	if v := viper.GetString("pikpak.api_server"); v != "" {
		apiServer = v
	}

	if v := viper.GetString("pikpak.device_id"); v != "" {
		deviceID = v
		resCli = resCli.SetHeader("X-Device-Id", deviceID)
	}
	if v := viper.GetString("pikpak.user_agent"); v != "" {
		userAgent = v
		resCli = resCli.SetHeader("User-Agent", userAgent)
	}

	autoUpdateCreateLinkLimitOptions(api, q)
	autoGetCreateLinkLimitOptions(api, q)
	go api.updatePremiumExpirationBackground()
	return api
}

// autoGetCreateLinkLimitList auto get create link limit options.
func autoGetCreateLinkLimitOptions(api *API, q *query.Query) {
	fn := func() {
		createLinkLimitOptions := make(map[string]createLinkLimitData)
		optionsUidInConfiguration := make([]string, 0)
		// The options in the configuration file take precedence
		if v := viper.GetStringMap("pikpak.create_link_limit"); len(v) > 0 {
			for keepshareUid, _ := range v {
				optionsUidInConfiguration = append(optionsUidInConfiguration, keepshareUid)
				createLinkLimitOptions[keepshareUid] = createLinkLimitData{
					IpNum:    viper.GetUint32(fmt.Sprintf("pikpak.create_link_limit.%s.ip_num", keepshareUid)),
					UnitTime: viper.GetUint32(fmt.Sprintf("pikpak.create_link_limit.%s.unit_time", keepshareUid)),
				}
			}
		}

		var result []*model.CreateLinkLimit
		err := q.CreateLinkLimit.Where(q.CreateLinkLimit.KeepshareUserID.NotIn(optionsUidInConfiguration...)).
			FindInBatches(
				&result, 100, func(tx gen.Dao, batch int) error {
					for _, v := range result {
						createLinkLimitOptions[v.KeepshareUserID] = createLinkLimitData{
							IpNum:    uint32(v.IPLimit),
							UnitTime: uint32(v.UnitTime),
						}
					}
					return nil
				},
			)
		if err != nil {
			log.Errorf("autoGetCreateLinkLimitOptions error: %v", err)
			return
		}
		if len(createLinkLimitOptions) > 0 {
			api.createLinkLimitOptions = createLinkLimitOptions
		}
	}
	fn()

	go func() {
		for {
			time.Sleep(time.Minute)
			fn()
		}
	}()
}

// autoUpdateCreateLinkLimitOptions auto update create link limit options.
func autoUpdateCreateLinkLimitOptions(api *API, q *query.Query) {
	type result struct {
		MasterUserID string `gorm:"column:master_user_id"`
		Total        int32  `gorm:"column:t"`
	}
	var baseLimitNum int32 = 2000
	fn := func() {
		var results []result
		pf := q.File
		err := api.Dependencies.Mysql.Table(pf.TableName()).
			Select(pf.MasterUserID.ColumnName().String(), "count(*) as t").
			Where(q.File.Status.Eq("PHASE_TYPE_RUNNING")).
			Group(pf.Status.ColumnName().String()).
			Group(pf.MasterUserID.ColumnName().String()).
			Having("t > ?", baseLimitNum).
			Find(&results).
			Error
		if err != nil {
			log.Errorf("autoUpdateCreateLinkLimitOptions error: %v", err)
			return
		}

		shouldLimitMasterIds := make([]string, 0)
		shouldLimitMapData := make(map[string]result)
		for _, v := range results {
			if v.Total > baseLimitNum {
				shouldLimitMasterIds = append(shouldLimitMasterIds, v.MasterUserID)
				shouldLimitMapData[v.MasterUserID] = v
			}
		}

		// query keepshare user id
		found, err := q.MasterAccount.Where(q.MasterAccount.UserID.In(shouldLimitMasterIds...)).Find()
		if err != nil {
			log.Errorf("autoUpdateCreateLinkLimitOptions find keepshare uid error: %v", err)
			return
		}

		now := time.Now()
		limitOptions := make([]*model.CreateLinkLimit, 0)
		for _, v := range found {
			r, ok := shouldLimitMapData[v.UserID]
			ipLimit := min(10, r.Total/1000)
			if ok && ipLimit > 0 {
				limitOptions = append(limitOptions, &model.CreateLinkLimit{
					KeepshareUserID: v.KeepshareUserID,
					IPLimit:         ipLimit,
					UnitTime:        int32(time.Hour.Seconds()) * 24 * 7,
					CreatedAt:       now,
					UpdatedAt:       now,
				})
			}
		}

		if len(limitOptions) > 0 {
			err := q.CreateLinkLimit.Clauses(clause.OnConflict{UpdateAll: true}).Create(limitOptions...)
			if err != nil {
				log.Errorf("autoUpdateCreateLinkLimitOptions create limit options error: %v", err)
			}
		}
		shouldLimitKeepShareUserIds := lo.Map(limitOptions, func(item *model.CreateLinkLimit, index int) string { return item.KeepshareUserID })
		q.CreateLinkLimit.Where(q.CreateLinkLimit.KeepshareUserID.NotIn(shouldLimitKeepShareUserIds...)).Delete()
	}
	fn()

	go func() {
		for {
			time.Sleep(time.Minute * 5)
			fn()
		}
	}()
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

func referralURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}
	return fmt.Sprintf("%s/%s", referralServer, strings.TrimLeft(path, "/"))
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
	passwordChars = []byte(`ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~!@#$%^&_+-=.`)
)

func randomPassword() string {
	const length = 12
	b := make([]byte, length)
	b[0] = passwordChars[apiRand.Intn(26)]   // start with a random uppercase letter
	b[1] = strconv.Itoa(apiRand.Intn(10))[0] // then a random number
	for i := 2; i < length; i++ {            // then random letters
		n := apiRand.Intn(len(passwordChars))
		b[i] = passwordChars[n]
	}
	return string(b)
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

// GetCreateLinkLimitList Gets create link restriction configuration data
func (api *API) GetCreateLinkLimitList() map[string]createLinkLimitData {
	return api.createLinkLimitOptions
}
