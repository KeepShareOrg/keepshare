package magnetblock

import (
	"context"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"

	"github.com/KeepShareOrg/keepshare/server/model"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var ErrBlocked = errors.New("magnet_blocked")

var errInvalidMagnet = errors.New("invalid magnet: expected one unambiguous BTIH info hash")

// Normalize returns the lowercase hexadecimal BTIH, independent of display name or trackers.
func Normalize(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil || !strings.EqualFold(u.Scheme, "magnet") || u.Opaque != "" || u.Host != "" || u.Path != "" || u.Fragment != "" {
		return "", errInvalidMagnet
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", errInvalidMagnet
	}
	var hash string
	for key, topics := range values {
		if !strings.EqualFold(key, "xt") {
			continue
		}
		for _, topic := range topics {
			if strings.EqualFold(topic, "urn:btih") {
				return "", errInvalidMagnet
			}
			if len(topic) < len("urn:btih:") || !strings.EqualFold(topic[:len("urn:btih:")], "urn:btih:") {
				continue
			}
			candidate, ok := normalizeHash(topic[len("urn:btih:"):])
			if !ok || (hash != "" && hash != candidate) {
				return "", errInvalidMagnet
			}
			hash = candidate
		}
	}
	if hash == "" {
		return "", errInvalidMagnet
	}
	return hash, nil
}

func normalizeHash(hash string) (string, bool) {
	var decoded []byte
	var err error
	switch len(hash) {
	case 40:
		decoded, err = hex.DecodeString(hash)
	case 32:
		decoded, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(hash))
	default:
		return "", false
	}
	if err != nil || len(decoded) != 20 {
		return "", false
	}
	return hex.EncodeToString(decoded), true
}

// Aliases returns both historical original_link_hash representations of a validated BTIH.
func Aliases(hash string) []string {
	decoded, err := hex.DecodeString(hash)
	if err != nil || len(decoded) != 20 {
		return nil
	}
	return []string{strings.ToLower(hash), strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(decoded))}
}

// IsBlocked ignores non-magnet links but fails closed on malformed magnets and database errors.
func IsBlocked(ctx context.Context, db *gorm.DB, link string) (bool, error) {
	if len(link) < len("magnet:") || !strings.EqualFold(link[:len("magnet:")], "magnet:") {
		return false, nil
	}
	hash, err := Normalize(link)
	if err != nil {
		return false, err
	}
	return IsHashBlocked(ctx, db, hash)
}

func IsHashBlocked(ctx context.Context, db *gorm.DB, hash string) (bool, error) {
	blocked, err := FindBlocked(ctx, db, []string{hash})
	return blocked[hash], err
}

// FindBlocked always reads the primary database; no process-local allow cache is used.
func FindBlocked(ctx context.Context, db *gorm.DB, hashes []string) (map[string]bool, error) {
	result := make(map[string]bool)
	originals := make(map[string][]string)
	canonical := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		normalized, ok := normalizeHash(hash)
		if !ok {
			continue
		}
		if _, exists := originals[normalized]; !exists {
			canonical = append(canonical, normalized)
		}
		originals[normalized] = append(originals[normalized], hash)
	}
	if len(canonical) == 0 {
		return result, nil
	}
	if db == nil {
		return nil, errors.New("magnet blacklist database is not configured")
	}
	var blocked []string
	err := db.WithContext(ctx).Clauses(dbresolver.Write).Model(&model.MagnetBlacklist{}).
		Where("info_hash IN ?", canonical).Pluck("info_hash", &blocked).Error
	if err != nil {
		return nil, err
	}
	for _, hash := range blocked {
		for _, original := range originals[hash] {
			result[original] = true
		}
	}
	return result, nil
}
