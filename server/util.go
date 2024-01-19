// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/KeepShareOrg/keepshare/config"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// VerifyRecaptchaToken verify google recaptcha token
func VerifyRecaptchaToken(token string) bool {
	secret := config.GoogleRecaptchaSecret()
	verifyApi := "https://www.google.com/recaptcha/api/siteverify"
	verifyUrl := fmt.Sprintf("%s?secret=%s&response=%s", verifyApi, secret, token)
	resp, err := http.Post(verifyUrl, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	data := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false
	}
	return data["success"].(bool)
}

func makeKeepSharingLink(channel, originalLink string) string {
	return fmt.Sprintf("https://%s/%s/%s", config.RootDomain(), channel, url.QueryEscape(originalLink))
}

func getOriginalLinks(src []string) (original, invalid []string) {
	original = make([]string, 0, len(src))
	for _, raw := range src {
		raw = strings.TrimSpace(raw)

		// check url
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" {
			invalid = append(invalid, raw)
			continue
		}

		// is auto sharing link
		if u.Host == config.RootDomain() {
			_, link, ok := getChannelAndLinkFromURL(u)
			if !ok {
				invalid = append(invalid, raw)
				continue
			}
			raw = link
		}

		original = append(original, lk.Simplify(raw))
	}
	return
}

// CalcSha265Hash calc hash with sha256
func CalcSha265Hash(input string, secret string) string {
	hash := sha256.New()
	content := fmt.Sprintf("%v%v", input, secret)

	hash.Write([]byte(content))
	hashedPassword := hex.EncodeToString(hash.Sum(nil))

	return hashedPassword
}

// GenerateVerificationCode generate random verification code
func GenerateVerificationCode(length int) string {
	if length <= 0 {
		length = 6
	}

	code := ""
	for i := 0; i < length; i++ {
		digit := rand.Intn(10)
		code += fmt.Sprint(digit)
	}

	return code
}

// GetRequestIP get request ip
func GetRequestIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

// RecordLinkAccessLog record link access log
func RecordLinkAccessLog(ctx context.Context, originalLinkHash, ip string) ([]string, error) {
	accessLog, err := config.Redis().Get(ctx, originalLinkHash).Result()
	if err != nil {
		accessLog = "[]"
	}

	var accessInfos []string
	_ = json.Unmarshal([]byte(accessLog), &accessInfos)

	log.WithContext(ctx).Debugf("access log: %#v", accessInfos)

	newAccessInfos := []string{
		fmt.Sprintf("%s-%d", ip, time.Now().Unix()),
	}

	RecordAccessIPNums := 10
	for _, v := range accessInfos {
		ipAndAccessTimestamp := strings.Split(v, "-")
		if len(ipAndAccessTimestamp) == 2 {
			accessTimestamp, err := strconv.Atoi(ipAndAccessTimestamp[1])
			if err != nil {
				continue
			}

			if time.Now().Sub(time.Unix(int64(accessTimestamp), 0)).Hours() < 24 && len(newAccessInfos) < RecordAccessIPNums {
				ipExist := false
				for _, v := range newAccessInfos {
					if strings.HasPrefix(v, ipAndAccessTimestamp[0]) {
						ipExist = true
					}
				}
				if !ipExist {
					newAccessInfos = append(newAccessInfos, v)
				}
			}
		}
	}

	if newAccessLog, err := json.Marshal(newAccessInfos); err == nil {
		_ = config.Redis().Set(ctx, originalLinkHash, string(newAccessLog), 24*time.Hour).Err()
	}

	return newAccessInfos, nil
}
