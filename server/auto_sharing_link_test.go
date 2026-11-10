package server

import (
	"context"
	"encoding/base32"
	"encoding/hex"
	"net/url"
	"strings"
	"testing"

	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/model"
)

func TestAutoSharingMagnetUsesOneCanonicalResource(t *testing.T) {
	const hash = "0123456789abcdef0123456789abcdef01234567"
	decoded, _ := hex.DecodeString(hash)
	encoded := base32.StdEncoding.EncodeToString(decoded)
	canonical := "magnet:?xt=urn:btih:" + hash
	for name, raw := range map[string]string{
		"unescaped hex":              "https://keepshare.invalid/deadbeef/" + canonical + "&dn=example&tr=https://tracker.invalid/announce",
		"escaped base32":             "https://keepshare.invalid/deadbeef/" + url.QueryEscape("magnet:?xt=urn:btih:"+encoded),
		"other topic first":          "https://keepshare.invalid/deadbeef/magnet:?xt=urn:sha1:example&xt=urn:btih:" + hash,
		"uppercase scheme and topic": "https://keepshare.invalid/deadbeef/MAGNET:?XT=URN:BTIH:" + strings.ToUpper(hash),
	} {
		t.Run(name, func(t *testing.T) {
			u, err := url.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			channel, link, ok := getChannelAndLinkFromURL(u)
			if !ok || channel != "deadbeef" {
				t.Fatalf("parse = %q, %q, %v", channel, link, ok)
			}
			simple, gotHash, ok := validateLink(link)
			if !ok || simple != canonical || gotHash != hash {
				t.Fatalf("validate = %q, %q, %v", simple, gotHash, ok)
			}
		})
	}
}

func TestAutoSharingMagnetRejectsConflictingTopics(t *testing.T) {
	const first = "0123456789abcdef0123456789abcdef01234567"
	magnet := "magnet:?xt=urn:btih:" + first + "&xt=urn:btih:" + strings.Repeat("0", 40)
	for _, path := range []string{magnet, url.QueryEscape(magnet)} {
		u, err := url.Parse("https://keepshare.invalid/deadbeef/" + path)
		if err != nil {
			t.Fatal(err)
		}
		_, link, ok := getChannelAndLinkFromURL(u)
		if !ok {
			t.Fatal("valid URL should reach magnet validation")
		}
		if _, _, ok := validateLink(link); ok {
			t.Fatal("conflicting BTIH topics must not be silently reduced to the first hash")
		}
	}
}

func TestPersonalBlockPrecedesHostStatusCache(t *testing.T) {
	record := &model.SharedLink{State: share.StatusBlocked.String(), HostSharedLink: "https://mypikpak.invalid/s/old"}
	if got := getShareStatus(context.Background(), "user", nil, record); got != share.StatusBlocked {
		t.Fatalf("blocked resource became %q", got)
	}
}

const (
	testEd2kHash = "8549fd92d11e47e3d9c3379e5ec8f842"
	testEd2k     = "ed2k://|file|The_Matrix.avi|735934464|8549FD92D11E47E3D9C3379E5EC8F842|/"
	testEd2kStr  = "ed2k://|file|The_Matrix.avi|735934464|" + testEd2kHash + "|/"
)

func TestValidateEd2k(t *testing.T) {
	simple, hash, ok := validateLink(testEd2k)
	if !ok || simple != testEd2kStr || hash != testEd2kHash {
		t.Fatalf("validateLink = %q, %q, %v", simple, hash, ok)
	}

	for _, raw := range []string{
		"ed2k://|server|1.2.3.4|4661|/",
		"ed2k://|file|a.bin|1024|8549fd92|/",
		"ed2k://|file|a.bin|-1|" + testEd2kHash + "|/",
	} {
		if _, _, ok := validateLink(raw); ok {
			t.Fatalf("validateLink(%q) must reject an invalid ed2k link", raw)
		}
	}
}

func TestAutoSharingEd2kUsesOneCanonicalResource(t *testing.T) {
	for name, path := range map[string]string{
		"unescaped": testEd2k,
		"escaped":   url.QueryEscape(testEd2k),
	} {
		t.Run(name, func(t *testing.T) {
			u, err := url.Parse("https://keepshare.invalid/deadbeef/" + path)
			if err != nil {
				t.Fatal(err)
			}
			channel, link, ok := getChannelAndLinkFromURL(u)
			if !ok || channel != "deadbeef" {
				t.Fatalf("parse = %q, %q, %v", channel, link, ok)
			}
			simple, hash, ok := validateLink(link)
			if !ok || simple != testEd2kStr || hash != testEd2kHash {
				t.Fatalf("validate = %q, %q, %v", simple, hash, ok)
			}
		})
	}
}
