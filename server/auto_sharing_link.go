// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/async"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gen"
)

var channelIDPattern = regexp.MustCompile(`[a-z0-9]{8}`)

func autoSharingLink(c *gin.Context) {
	channel, link, ok := getChannelAndLinkFromURL(c.Request.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_request", i18n.WithDataMap("error", "invalid auto sharing url")))
		return
	}
	if !channelIDPattern.MatchString(channel) {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_channel", i18n.WithDataMap("channel", channel)))
		return
	}
	linkRaw, _, ok := validateLink(link)
	if !ok {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_link", i18n.WithDataMap("link", link)))
		return
	}

	ctx := c.Request.Context()
	hostName := util.FirstNotEmpty(c.Query("host"), config.DefaultHost())
	host := hosts.Get(hostName)
	if host == nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_host", i18n.WithDataMap("host", hostName)))
		return
	}

	user, err := query.User.WithContext(ctx).Where(query.User.Channel.Eq(channel)).Take()
	if err != nil && !gormutil.IsNotFoundError(err) {
		mdw.RespInternal(c, err.Error())
		return
	}
	if user == nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_channel", i18n.WithDataMap("channel", channel)))
		return
	}

	l := log.WithFields(Map{
		constant.UserID: user.ID,
		"link":          linkRaw,
		"host":          hostName,
	})

	sh, err := createShareLinkIfNotExist(c, user.ID, host, link, share.AutoShare)
	if err != nil {
		mdw.RespInternal(c, err.Error())
		return
	}

	l = l.WithFields(Map{constant.SharedLink: sh.HostSharedLink, constant.ShareStatus: sh.State})

	switch share.State(sh.State) {
	case share.StatusOK:
		l.Debug("got shared_link")
		c.Redirect(http.StatusFound, sh.HostSharedLink)

	default: // include StatusSensitive
		l.Debug("share status:", sh.State)
		statusPageAddress := fmt.Sprintf("https://%s/console/shared/status?id=%v", config.RootDomain(), sh.AutoID)
		c.Redirect(http.StatusFound, statusPageAddress)
	}
}

// createShareLinkIfNotExist if the shared link does not exist, create a new one and return it.
func createShareLinkIfNotExist(ctx *gin.Context, userID string, host *hosts.HostWithProperties, link string, createBy string) (*model.SharedLink, error) {
	linkRaw, linkHash, ok := validateLink(link)
	if !ok || linkHash == "" {
		return nil, errors.New("invalid link")
	}

	where := []gen.Condition{
		query.SharedLink.UserID.Eq(userID),
		query.SharedLink.OriginalLinkHash.Eq(linkHash),
	}

	sh, err := query.SharedLink.WithContext(ctx).Where(where...).Take()
	if err != nil && !gormutil.IsNotFoundError(err) {
		return nil, fmt.Errorf("query shared link error: %w", err)
	}

	if sh != nil {
		status := getShareStatus(ctx, userID, host, sh)
		switch status {
		case share.StatusUnknown, share.StatusOK, share.StatusCreated, share.StatusPending:
			break

		case share.StatusDeleted, share.StatusNotFound, share.StatusSensitive:
			sh = nil // re-create a shared link

		case share.StatusBlocked:
			return nil, errors.New("link_blocked")
		default:
			return nil, fmt.Errorf("unexpected share status: %s", status)
		}
	}

	if sh == nil {
		sh, err = createShareByLink(ctx, userID, host, linkRaw, createBy)
		if err != nil {
			return nil, fmt.Errorf("create share error: %w", err)
		}
	}

	return sh, nil
}

func getChannelAndLinkFromURL(u *url.URL) (channel, link string, ok bool) {
	path := strings.TrimPrefix(u.Path, "/")
	channel, link, _ = strings.Cut(path, "/")
	channel = strings.ToLower(channel)
	link = strings.TrimSpace(strings.TrimLeft(link, "/"))
	if channel == "" || link == "" {
		return "", "", false
	}

	// compatible for not url encoded magnet
	if link == "magnet:" {
		link = link + "?xt=" + u.Query().Get("xt")
	}

	u, err := url.Parse(link)
	if err != nil || u.Scheme == "" {
		return "", "", false
	}

	return channel, link, true
}

func splitPath(path string) (channel, link string) {
	path = strings.TrimPrefix(path, "/")
	channel, link, _ = strings.Cut(path, "/")
	channel = strings.ToLower(channel)
	return
}

func validateLink(link string) (simple string, hash string, ok bool) {
	link = strings.TrimLeft(link, "/")
	link = strings.TrimSpace(link)

	if len(link) == 0 {
		return "", "", false
	}

	u, _ := url.Parse(link)
	if u == nil || u.Scheme == "" {
		return "", "", false
	}

	simple = lk.Simplify(link)
	hash = lk.Hash(link)

	if hash == "" {
		return "", "", false
	}

	return simple, hash, true
}

func createShareByLink(ctx context.Context, userID string, host *hosts.HostWithProperties, link string, createBy string) (s *model.SharedLink, err error) {
	sharedLinks, err := host.CreateFromLinks(ctx, userID, []string{link}, createBy)
	if err != nil {
		return nil, fmt.Errorf("create share from links err: %w", err)
	}

	sh := sharedLinks[link]
	if sh == nil {
		return nil, errors.New("get nil share")
	}

	now := time.Now()
	s = &model.SharedLink{
		UserID:             userID,
		State:              sh.State.String(),
		Host:               host.Name(),
		CreatedBy:          sh.CreatedBy,
		CreatedAt:          now,
		UpdatedAt:          now,
		Size:               sh.Size,
		Visitor:            sh.Visitor,
		Stored:             sh.Stored,
		Revenue:            sh.Revenue,
		Title:              sh.Title,
		OriginalLinkHash:   lk.Hash(link),
		HostSharedLinkHash: lk.Hash(sh.HostSharedLink),
		OriginalLink:       link,
		HostSharedLink:     sh.HostSharedLink,
	}
	if createBy == share.AutoShare {
		s.LastVisitedAt = now
	}

	if err = query.SharedLink.Create(s); err != nil {
		err = fmt.Errorf("create shared record err: %w", err)
		log.WithField("shared_record", s).Error(err)
		return nil, err
	}
	log.WithField("shared_record", s).Info("create shared record done")

	return s, nil
}

// ignore errors.
func getShareStatus(ctx context.Context, userID string, host hosts.Host, record *model.SharedLink) share.State {
	const statusCacheTTL = time.Minute

	id := record.AutoID
	link := record.HostSharedLink
	key := fmt.Sprintf("status:%d", id)

	cache, _ := config.Redis().Get(ctx, key).Result()
	if cache != "" {
		return share.State(cache)
	}

	sts, err := host.GetStatuses(ctx, userID, []string{link})
	if err != nil {
		log.WithFields(Map{
			constant.SharedLink: link,
			constant.Error:      err,
			constant.UserID:     userID,
		}).Error("get status error")
		return share.StatusUnknown
	}

	status, ok := sts[link]
	if !ok || status == "" {
		status = share.StatusUnknown
	}

	// update state and last visit time.
	now := time.Now()
	_, _ = query.SharedLink.WithContext(ctx).Updates(&model.SharedLink{
		AutoID:        id,
		State:         status.String(),
		LastVisitedAt: now,
		UpdatedAt:     now,
	})

	async.Run(func() { getStatisticsLater(id) })

	if status == share.StatusOK {
		config.Redis().Set(ctx, key, status.String(), statusCacheTTL)
	}
	return status
}
