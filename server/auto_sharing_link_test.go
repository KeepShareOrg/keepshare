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
