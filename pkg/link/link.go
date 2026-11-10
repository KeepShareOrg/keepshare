// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package link

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	// MagnetPrefix is the prefix of magnet links.
	MagnetPrefix = "magnet:?"

	// Ed2kPrefix is the prefix of ed2k links.
	Ed2kPrefix = "ed2k://"

	ed2kHashLength = 32
)

// Ed2kFile is the parsed representation of an ed2k file link.
type Ed2kFile struct {
	// Name is the file name as it appears in the link.
	Name string
	// Size is the file size in bytes.
	Size int64
	// Hash is the lowercase hexadecimal MD4 hash of the file content.
	Hash string
}

// String returns the canonical ed2k file link of the given file.
func (f *Ed2kFile) String() string {
	return fmt.Sprintf("%s|file|%s|%d|%s|/", Ed2kPrefix, f.Name, f.Size, f.Hash)
}

// IsEd2k reports whether link is an ed2k link.
func IsEd2k(link string) bool {
	return len(link) >= len(Ed2kPrefix) && strings.EqualFold(link[:len(Ed2kPrefix)], Ed2kPrefix)
}

// ParseEd2k parses an ed2k file link in the form
// ed2k://|file|<name>|<size>|<hash>|[h=<aich>|]/ .
// Server and serverlist links are not accepted.
func ParseEd2k(raw string) (*Ed2kFile, bool) {
	raw = strings.TrimSpace(raw)
	if !IsEd2k(raw) {
		return nil, false
	}

	fields := strings.Split(raw[len(Ed2kPrefix):], "|")
	// fields: ["", file, name, size, hash, "/"] or
	// ["", file, name, size, hash, "h=<aich>", "/"]
	if len(fields) < 6 || fields[0] != "" || !strings.EqualFold(fields[1], "file") || fields[len(fields)-1] != "/" {
		return nil, false
	}
	if len(fields) != 6 && (len(fields) != 7 || !strings.HasPrefix(fields[5], "h=") || len(fields[5]) == 2) {
		return nil, false
	}

	name := fields[2]
	if name == "" {
		return nil, false
	}
	size, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil || size < 0 {
		return nil, false
	}
	hash := strings.ToLower(fields[4])
	if len(hash) != ed2kHashLength {
		return nil, false
	}
	if _, err = hex.DecodeString(hash); err != nil {
		return nil, false
	}

	return &Ed2kFile{Name: name, Size: size, Hash: hash}, true
}

// Simplify this link and remove useless content.
// Keep only the xt parameter of the magnet link, or the file fields of the ed2k link.
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

	case IsEd2k(raw):
		if f, ok := ParseEd2k(raw); ok {
			return f.String()
		}
		return raw

	default:
		return raw
	}
}

var infoHashPattern = regexp.MustCompile("^[0-9a-zA-Z]+")

// Hash the magnet link's hash is infoHash, the ed2k link's hash is the ed2k
// file hash, other link's hash is the sha1 hash of the link, in lowercase.
// Invalid magnet and ed2k links have no hash.
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

	if IsEd2k(link) {
		if f, ok := ParseEd2k(link); ok {
			return f.Hash
		}
		return ""
	}

	h := sha1.Sum([]byte(link))
	return fmt.Sprintf("%x", h)
}

func isMagnet(link string) bool {
	return len(link) > len(MagnetPrefix) && strings.EqualFold(link[:len(MagnetPrefix)], MagnetPrefix)
}
