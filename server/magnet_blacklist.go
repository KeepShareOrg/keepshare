package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	pikpakapi "github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	pikpakmodel "github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/magnetblock"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sharedLinkResponse struct {
	*model.SharedLink
	ResourceOnly bool `json:"resource_only"`
}

// Apply the policy at every response boundary, including historical records.
func sharedLinkResponses(ctx context.Context, records []*model.SharedLink, infoOnly bool) ([]*sharedLinkResponse, error) {
	hashes := make([]string, len(records))
	for i, record := range records {
		if isMagnetResource(record.OriginalLink) {
			hash, err := magnetblock.Normalize(record.OriginalLink)
			if err != nil {
				continue
			}
			hashes[i] = hash
		}
	}
	blocked, err := magnetblock.FindBlocked(ctx, config.MySQL(), hashes)
	if err != nil {
		return nil, err
	}
	responses := make([]*sharedLinkResponse, 0, len(records))
	for i, record := range records {
		denied := blocked[hashes[i]] || (isMagnetResource(record.OriginalLink) && hashes[i] == "")
		response := &sharedLinkResponse{SharedLink: record, ResourceOnly: infoOnly || denied}
		if denied {
			record.State = share.StatusBlocked.String()
		}
		if response.ResourceOnly || record.State == share.StatusBlocked.String() {
			record.HostSharedLink = ""
			record.HostSharedLinkHash = ""
			record.Error = ""
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func isMagnetResource(link string) bool {
	return len(link) >= len("magnet:") && strings.EqualFold(link[:len("magnet:")], "magnet:")
}

// This path deliberately does not consult a host or start a creation task.
func resourceSharedLink(ctx context.Context, userID, host, link string) (*model.SharedLink, error) {
	hash := lk.Hash(link)
	aliases := resourceHashAliases(link)
	t := query.SharedLink
	find := func() (*model.SharedLink, error) {
		return t.WithContext(ctx).Where(t.UserID.Eq(userID), t.Host.Eq(host), t.OriginalLinkHash.In(aliases...)).Take()
	}
	record, err := find()
	if err == nil {
		return record, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	now := time.Now()
	record = &model.SharedLink{
		UserID: userID, Host: host, State: share.StatusCreated.String(),
		CreatedBy: share.AutoShare, CreatedAt: now, UpdatedAt: now, LastVisitedAt: now,
		OriginalLink: link, OriginalLinkHash: hash,
	}
	if err := t.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(record); err != nil {
		return nil, err
	}
	return find()
}

func resourceHashAliases(link string) []string {
	if isMagnetResource(link) {
		if hash, err := magnetblock.Normalize(link); err == nil {
			return magnetblock.Aliases(hash)
		}
	}
	return []string{lk.Hash(link)}
}

// A stale callback must not resurrect either a global or a personal block.
func writableSharedLinks(db *gorm.DB, userID, host, hash string) *gorm.DB {
	aliases := []string{hash}
	canonical, err := magnetblock.Normalize("magnet:?xt=urn:btih:" + hash)
	if err == nil {
		aliases = magnetblock.Aliases(canonical)
	}
	db = db.Model(&model.SharedLink{}).
		Where("user_id = ? AND host = ? AND original_link_hash = ?", userID, host, hash).
		Where("state <> ?", share.StatusBlocked.String()).
		Where("NOT EXISTS (SELECT 1 FROM keepshare_blacklist WHERE user_id = ? AND original_link_hash IN ?)", userID, aliases)
	if err == nil {
		db = db.Where("NOT EXISTS (SELECT 1 FROM keepshare_magnet_blacklist WHERE info_hash = ?)", canonical)
	}
	return db
}

// Serialize publication with revocation's mapping lock, including after a block is removed.
func updateSharedLink(db *gorm.DB, userID, host, hash string, update *model.SharedLink) error {
	if host != "pikpak" || update.HostSharedLink == "" {
		return writableSharedLinks(db, userID, host, hash).Updates(update).Error
	}
	shareID, err := pikpakapi.ShareIDFromLink(update.HostSharedLink)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var native pikpakmodel.SharedLink
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("share_id = ?", shareID).Take(&native).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // A delayed callback must not publish a revoked URL.
		}
		if err != nil {
			return err
		}
		return writableSharedLinks(tx, userID, host, hash).Updates(update).Error
	})
}

func statusPageURL(path string, autoID int64, requestID string) string {
	return fmt.Sprintf("https://%s/console/shared/%s?id=%d&request_id=%s", config.RootDomain(), path, autoID, url.QueryEscape(requestID))
}
