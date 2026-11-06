package magnetblock

import (
	"context"
	"encoding/base32"
	"encoding/hex"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	const hash = "0123456789abcdef0123456789abcdef01234567"
	bytes, _ := hex.DecodeString(hash)
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	for name, link := range map[string]string{
		"hex":               "magnet:?xt=urn:btih:" + hash,
		"base32":            "magnet:?xt=urn:btih:" + encoded,
		"lowercase base32":  "magnet:?xt=urn:btih:" + strings.ToLower(encoded),
		"case and metadata": "MAGNET:?XT=URN:BTIH:" + strings.ToUpper(hash) + "&dn=Another+name&tr=https%3A%2F%2Ftracker.example%2Fannounce",
		"other xt type":     "magnet:?xt=urn:sha1:example&xt=urn:btih:" + hash,
		"equivalent topics": "magnet:?xt=urn:btih:" + hash + "&xt=urn:btih:" + encoded,
		"repeated topic":    "magnet:?xt=urn:btih:" + hash + "&xt=urn:btih:" + hash,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := Normalize(link)
			if err != nil || got != hash {
				t.Fatalf("Normalize() = %q, %v; want %q", got, err, hash)
			}
		})
	}
}

func TestNormalizeRejectsInvalidAndAmbiguousTopics(t *testing.T) {
	const hash = "0123456789abcdef0123456789abcdef01234567"
	for name, link := range map[string]string{
		"empty":                "",
		"not magnet":           "https://example.com/?xt=urn:btih:" + hash,
		"no topic":             "magnet:?dn=example",
		"other topic only":     "magnet:?xt=urn:sha1:example",
		"empty hash":           "magnet:?xt=urn:btih:",
		"short hex":            "magnet:?xt=urn:btih:" + hash[:39],
		"invalid hex":          "magnet:?xt=urn:btih:" + strings.Repeat("g", 40),
		"invalid base32":       "magnet:?xt=urn:btih:" + strings.Repeat("0", 32),
		"padded base32":        "magnet:?xt=urn:btih:" + strings.Repeat("A", 32) + "=",
		"conflicting topics":   "magnet:?xt=urn:btih:" + hash + "&xt=urn:btih:" + strings.Repeat("0", 40),
		"malformed extra btih": "magnet:?xt=urn:btih:" + hash + "&xt=urn:btih:bad",
		"missing separator":    "magnet:?xt=urn:btih:" + hash + "&xt=urn:btih",
		"invalid escaping":     "magnet:?xt=urn:btih:" + hash + "&dn=%zz",
		"opaque uri":           "magnet:urn:btih:" + hash,
		"unexpected host":      "magnet://host?xt=urn:btih:" + hash,
	} {
		t.Run(name, func(t *testing.T) {
			if got, err := Normalize(link); err == nil {
				t.Fatalf("Normalize() = %q without error", got)
			}
		})
	}
}

func TestAliases(t *testing.T) {
	const hash = "0000000000000000000000000000000000000000"
	aliases := Aliases(hash)
	if len(aliases) != 2 || aliases[0] != hash || aliases[1] != strings.Repeat("a", 32) {
		t.Fatalf("Aliases() = %v", aliases)
	}
	for _, alias := range aliases {
		got, err := Normalize("magnet:?xt=urn:btih:" + alias)
		if err != nil || got != hash {
			t.Fatalf("alias %q normalizes to %q, %v", alias, got, err)
		}
	}
}

func TestUnrelatedInputsDoNotRequireDatabase(t *testing.T) {
	ctx := context.Background()
	for _, link := range []string{"", "https://example.com/file", "ed2k://example"} {
		if blocked, err := IsBlocked(ctx, nil, link); blocked || err != nil {
			t.Fatalf("IsBlocked(%q) = %v, %v", link, blocked, err)
		}
	}
	if blocked, err := IsHashBlocked(ctx, nil, "invalid"); blocked || err != nil {
		t.Fatalf("IsHashBlocked(invalid) = %v, %v", blocked, err)
	}
	if found, err := FindBlocked(ctx, nil, []string{"", "invalid"}); len(found) != 0 || err != nil {
		t.Fatalf("FindBlocked(invalid) = %v, %v", found, err)
	}
	if _, err := IsBlocked(ctx, nil, "magnet:?xt=urn:btih:bad"); err == nil {
		t.Fatal("malformed magnet must return an error")
	}
	if _, err := IsHashBlocked(ctx, nil, strings.Repeat("0", 40)); err == nil {
		t.Fatal("valid hash must not fail open without a database")
	}
}
