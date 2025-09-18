// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/async"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/model"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

var channelIDPattern = regexp.MustCompile(`[a-z0-9]{8}`)

const (
	keyHostLink     = "host_link"
	keyRedirectType = "redirect_type"
	keyState        = "state"
	keyTotalMS      = "total_ms"
)

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

	requestID, _ := log.RequestIDFromContext(ctx)
	fields := Map{
		constant.IP:        c.ClientIP(),
		constant.DeviceID:  c.GetHeader(constant.HeaderDeviceID),
		constant.RequestID: requestID,
		constant.UserID:    user.ID,
		constant.Channel:   channel,
		constant.Link:      linkRaw,
		constant.Host:      hostName,
	}

	log.ContextWithFields(ctx, fields)
	report := log.NewReport("visit_link").Sets(fields).Sets(Map{
		keyHostLink:     "error",
		keyRedirectType: "error",
		keyState:        "error",
	})
	defer report.Done()

	shouldRedirectToWhatsLinkInfoPage := false
	if c.Query("wsl") != "" {
		shouldRedirectToWhatsLinkInfoPage = true
	}
	warningChannels := viper.GetStringSlice("warning_channels")
	if warningChannels != nil {
		log.Debugf("warning channels: %v", channel)
		shouldRedirectToWhatsLinkInfoPage = slices.Contains(warningChannels, channel)
	}

	l := log.WithContext(ctx)
	shouldSkipCreateLink := shouldRedirectToWhatsLinkInfoPage
	// if the channel and the link are warning, forbid to create shared link
	hitForbidden := checkForbiddenRules(ctx, channel, link)
	if hitForbidden {
		shouldSkipCreateLink = true
	}

	ctx = context.WithValue(ctx, constant.IsShouldSkipCreateLink, shouldSkipCreateLink)
	sh, lastState, err := createShareLinkIfNotExist(ctx, user.ID, host, link, share.AutoShare, c.ClientIP())
	if err != nil {
		report.Set(constant.Error, err.Error())
		mdw.RespInternal(c, err.Error())
		return
	}

	report.Set(keyState, lastState)
	l = l.WithFields(Map{constant.SharedLink: sh.HostSharedLink, constant.ShareStatus: sh.State})

	// if the link refer to the warning channel id, we need redirect to the whatslink info page
	if shouldRedirectToWhatsLinkInfoPage {
		l.Debug("redirect to whatslink info page")
		c.Redirect(http.StatusFound, fmt.Sprintf("https://%s/console/shared/wsl-status?id=%d&request_id=%s", config.RootDomain(), sh.AutoID, requestID))
		return
	}

	switch share.State(sh.State) {
	case share.StatusOK:
		// We can add parameters to the PikPak sharing page to automatically play, but we need to add that only the current host is PikPak.
		hostLink := fmt.Sprintf("%s?act=play", sh.HostSharedLink)
		report.Sets(Map{
			keyRedirectType: "share",
			keyHostLink:     hostLink,
		})
		l.Debug("got shared_link")
		c.Redirect(http.StatusFound, hostLink)

	default: // include StatusSensitive
		l.Debug("share status:", sh.State)
		RecordLinkAccessLog(ctx, sh.OriginalLinkHash, GetRequestIP(c.Request))
		statusPage := fmt.Sprintf("https://%s/console/shared/status?id=%d&request_id=%s", config.RootDomain(), sh.AutoID, requestID)
		// skip the status page loading if not yet create host task
		if lastState == share.StatusCreated {
			// st: slow task, tasks that have been created for a while but have not yet been completed
			statusPage = fmt.Sprintf("%v&st=%d", statusPage, 1)
		}
		report.Sets(Map{
			keyRedirectType: "status",
			keyHostLink:     statusPage,
		})
		c.Redirect(http.StatusFound, statusPage)
	}
}

// checkForbiddenRules check if the channel and the link are forbidden
func checkForbiddenRules(ctx context.Context, channelID, link string) bool {
	type Rule struct {
		ChannelID       string   `mapstructure:"channel_id"`
		FilenameContain []string `mapstructure:"filename_contain"`
	}

	var rules []*Rule
	if err := viper.UnmarshalKey("forbidden_rules", &rules); err != nil {
		log.Errorf("failed to unmarshal forbidden_rules: %v", err)
		return false
	}

	rule, hit := lo.Find(rules, func(item *Rule) bool {
		return item.ChannelID == channelID
	})
	if !hit || rule == nil {
		return false
	}

	info, err := queryLinkInfoByWhatsLinks(ctx, link)
	if err != nil {
		log.Errorf("query what's link info error: %v", err)
		return false
	}
	hit = lo.SomeBy(rule.FilenameContain, func(item string) bool {
		return strings.Contains(info.Name, item)
	})

	return hit
}

type WhatsLinkInfo struct {
	Error       string `json:"error"`
	Type        string `json:"type"`
	FileType    string `json:"file_type"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	Count       int    `json:"count"`
	Screenshots []struct {
		Time       int    `json:"time"`
		Screenshot string `json:"screenshot"`
	} `json:"screenshots"`
}

func queryLinkInfoByWhatsLinks(ctx context.Context, link string) (*WhatsLinkInfo, error) {
	addr := fmt.Sprintf("https://whatslink.info/api/v1/link?url=%s", link)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bs, _ := io.ReadAll(resp.Body)
	log.Debugf("whats link return: %s", bs)

	var linkInfo *WhatsLinkInfo
	err = json.Unmarshal(bs, &linkInfo)
	if err != nil {
		return nil, err
	}

	return linkInfo, nil
}

// createShareLinkIfNotExist if the shared link does not exist, create a new one and return it.
func createShareLinkIfNotExist(ctx context.Context, userID string, host *hosts.HostWithProperties, link string, createBy string, ip string) (sharedLink *model.SharedLink, lastStatus share.State, err error) {
	linkRaw, linkHash, ok := validateLink(link)
	if !ok || linkHash == "" {
		return nil, "", errors.New("invalid link")
	}

	var sh *model.SharedLink
	if res, err := query.SharedLinkComplete.WithContext(ctx).Where(
		query.SharedLinkComplete.UserID.Eq(userID),
		query.SharedLinkComplete.OriginalLinkHash.Eq(linkHash),
	).Take(); err == nil {
		sh = (*model.SharedLink)(res)
	}
	if sh == nil {
		sh, err = query.SharedLink.WithContext(ctx).Where(
			query.SharedLink.UserID.Eq(userID),
			query.SharedLink.OriginalLinkHash.Eq(linkHash),
		).Take()
		if err != nil && !gormutil.IsNotFoundError(err) {
			return nil, "", fmt.Errorf("query shared link error: %w", err)
		}
	}

	lastStatus = share.StatusNotFound
	if sh != nil {
		lastStatus = getShareStatus(ctx, userID, host, sh)
		go updateVisitTimeAndState(ctx, sh, lastStatus)
		switch lastStatus {
		case share.StatusUnknown, share.StatusOK, share.StatusCreated, share.StatusPending:
			break

		case share.StatusDeleted, share.StatusNotFound, share.StatusSensitive, share.StatusError:
			sh = nil // re-create a shared link

		case share.StatusBlocked:
			return nil, lastStatus, errors.New("link_blocked")

		default:
			return nil, lastStatus, fmt.Errorf("unexpected share status: %s", lastStatus)
		}
	}

	if sh == nil {
		sh, err = createShareByLink(ctx, userID, host, linkRaw, createBy, ip)
		if err != nil {
			return nil, lastStatus, fmt.Errorf("create share error: %w", err)
		}
	}
	return sh, lastStatus, nil
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

func createShareByLink(ctx context.Context, userID string, host *hosts.HostWithProperties, link string, createBy string, ip string) (s *model.SharedLink, err error) {
	now := time.Now()
	s = &model.SharedLink{
		State:            string(share.StatusCreated),
		UserID:           userID,
		CreatedBy:        createBy,
		Host:             host.Name(),
		CreatedAt:        now,
		UpdatedAt:        now,
		OriginalLinkHash: lk.Hash(link),
		OriginalLink:     link,
	}
	if createBy == share.AutoShare {
		s.LastVisitedAt = now
	}

	t := query.SharedLink
	err = t.WithContext(ctx).
		Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{
			t.CreatedAt.ColumnName().String(),
			t.UpdatedAt.ColumnName().String(),
		})}).
		Create(s)
	if err != nil {
		err = fmt.Errorf("create shared record err: %w", err)
		log.WithContext(ctx).WithField("shared_record", s).Error(err)
		return nil, err
	}
	log.WithContext(ctx).WithField("shared_record", s).Info("create shared record done")

	isWarningChannel := ctx.Value(constant.IsShouldSkipCreateLink)
	if isWarningChannel != "true" {
		return s, nil
	}
	go func() {
		ctx = context.Background()
		sharedLinks, err := host.CreateFromLinks(ctx, userID, []string{link}, createBy, ip)
		if err != nil {
			log.WithContext(ctx).WithField("shared_record", s).Error(fmt.Errorf("create share from links err: %w", err))
			return
		}

		sh := sharedLinks[link]

		if sh == nil {
			log.WithContext(ctx).WithField("shared_record", s).Error(errors.New("get nil keepshare"))
			return
		}

		update := &model.SharedLink{
			State:              sh.State.String(),
			Size:               sh.Size,
			Visitor:            sh.Visitor,
			Stored:             sh.Stored,
			Revenue:            sh.Revenue,
			Title:              sh.Title,
			HostSharedLinkHash: lk.Hash(sh.HostSharedLink),
			HostSharedLink:     sh.HostSharedLink,
			Error:              sh.Error,
		}
		log.WithContext(ctx).WithField("shared_record", s).Infof("sharedLinks update :%+v", update)
		_, err = t.WithContext(ctx).Where(t.AutoID.Eq(s.AutoID)).Updates(update)
		if err != nil {
			log.WithContext(ctx).WithField("shared_record", s).WithField("autoID", s.AutoID).Error(errors.New("get nil share"))
			return
		}
	}()

	return s, nil
}

// ignore errors.
func getShareStatus(ctx context.Context, userID string, host hosts.Host, record *model.SharedLink) share.State {
	link := record.HostSharedLink
	if link == "" {
		return share.State(record.State)
	}

	const statusCacheTTL = time.Minute
	id := record.AutoID
	key := fmt.Sprintf("status:%d", id)

	cache, _ := config.Redis().Get(ctx, key).Result()
	if cache != "" {
		return share.State(cache)
	}

	sts, err := host.GetStatuses(ctx, userID, []string{link})
	if err != nil {
		log.WithContext(ctx).WithFields(Map{
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

	async.Run(func() { getStatisticsLater(id) })

	if status == share.StatusOK {
		config.Redis().Set(ctx, key, status.String(), statusCacheTTL)
	}
	return status
}

// update state and last visit time.
func updateVisitTimeAndState(ctx context.Context, record *model.SharedLink, status share.State) {
	now := time.Now()
	updates := &model.SharedLink{
		AutoID:        record.AutoID,
		LastVisitedAt: now,
		UpdatedAt:     now,
	}

	var needUpdate = now.Sub(record.LastVisitedAt) > time.Minute
	if status != "" && status != share.StatusUnknown && status != share.State(record.State) {
		updates.State = status.String()
		needUpdate = true
	}

	if !needUpdate {
		return
	}

	_, _ = query.SharedLink.WithContext(ctx).Updates(updates)
}
