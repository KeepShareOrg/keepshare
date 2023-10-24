// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package link

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	// MagnetPrefix is the prefix of magnet links.
	MagnetPrefix = "magnet:?"
)

// Simplify this link and remove useless content.
// Keep only the xt parameter of the magnet link.
func Simplify(raw string) (simple string) {
	defer func() {
		simple = strings.ToValidUTF8(simple, "")
	}()

	raw = strings.TrimSpace(raw)
	switch {
	case isMagnet(raw):
		v, _ := url.ParseQuery(raw[len(MagnetPrefix):])
		xt := strings.ToLower(v.Get("xt"))
		if xt == "" {
			return raw
		}
		return MagnetPrefix + "xt=" + xt

	default:
		return raw
	}
}

var infoHashPattern = regexp.MustCompile("^[0-9a-zA-Z]+")

// Hash the magnet link's hash is infoHash, other link's hash is the sha1 hash of the link, in lowercase
func Hash(link string) string {
	if isMagnet(link) {
		q, _ := url.ParseQuery(link[len(MagnetPrefix):])
		xt := strings.ToLower(q.Get("xt"))
		h := strings.TrimPrefix(xt, "urn:btih:")
		h = infoHashPattern.FindString(h)
		if len(h) > 40 {
			h = h[:40]
		}
		return strings.ToLower(h)
	}

	h := sha1.Sum([]byte(link))
	return fmt.Sprintf("%x", h)
}

func isMagnet(link string) bool {
	return len(link) > len(MagnetPrefix) && strings.EqualFold(link[:len(MagnetPrefix)], MagnetPrefix)
}
