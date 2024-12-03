// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/account"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	lk "github.com/KeepShareOrg/keepshare/pkg/link"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/samber/lo"
)

// CreateShare create a sharing link by files.
func (p *PikPak) CreateShare(ctx context.Context, master string, worker string, fileID string) (sharedLink string, err error) {
	return p.api.CreateShare(ctx, master, worker, fileID)
}

// CreateFromLinks create shared links based on the input original links.
func (p *PikPak) CreateFromLinks(ctx context.Context, keepShareUserID string, originalLinks []string, createBy string) (sharedLinks map[string]*share.Share, err error) {
	log := log.WithContext(ctx)
	defer func() {
		if err != nil {
			log.WithContext(ctx).Error("CreateFromLinks err:", err)
		}
	}()

	master, err := p.m.GetMaster(ctx, keepShareUserID)
	if err != nil {
		return nil, err
	}

	hashToLink := make(map[string]string, len(originalLinks))
	hashes := make([]string, 0, len(originalLinks))
	for _, v := range originalLinks {
		h := lk.Hash(v)
		hashes = append(hashes, h)
		hashToLink[h] = v
	}

	files, err := p.q.File.WithContext(ctx).Where(
		p.q.File.OriginalLinkHash.In(hashes...),
		p.q.File.MasterUserID.Eq(master.UserID),
	).Find()
	if err != nil && !gormutil.IsNotFoundError(err) {
		return nil, fmt.Errorf("query files err: %w", err)
	}

	linksStatusOK := map[string]*model.File{}
	linksStatusNotCompleted := map[string]*model.File{}
	linksStatusError := map[string]*model.File{}
	for _, f := range files {
		l := hashToLink[f.OriginalLinkHash]
		switch f.Status {
		case comm.StatusOK:
			linksStatusOK[l] = f
		case comm.StatusError:
			linksStatusError[l] = f
			// delete error files
			p.q.File.WithContext(ctx).Delete(f)
		default:
			linksStatusNotCompleted[l] = f
		}
	}

	sharedLinks = map[string]*share.Share{}
	// only create files for new links.
	for _, link := range originalLinks {
		if linksStatusOK[link] != nil || linksStatusError[link] != nil || linksStatusNotCompleted[link] != nil {
			continue
		}

		//if the link status is ok, it means the link will complete soon
		if createBy == share.AutoShare {
			status, progress := p.api.QueryLinkStatus(ctx, link)
			if status == comm.LinkStatusLimited {
				sh := &share.Share{
					State:        share.StatusSensitive,
					OriginalLink: link,
					CreatedBy:    createBy,
					CreatedAt:    time.Now(),
					Error:        "status:" + comm.LinkStatusLimited,
				}
				sharedLinks[link] = sh
				continue
			}

			if status != comm.LinkStatusOK && progress < 95 {
				/* because the machine is slow, we need to check the access info to limit the rate */
				//infos, err := GetLinkAccessInfos(ctx, lk.Hash(link))
				//if err == nil && len(infos) < comm.SlowTaskTriggerConditionTimes {
				//	sh := &share.Share{
				//		State:        share.StatusCreated,
				//		OriginalLink: link,
				//		CreatedBy:    createBy,
				//		CreatedAt:    time.Now(),
				//	}
				//	sharedLinks[link] = sh
				//	continue
				//}
			}
		}

		size, err := p.queryFilesizeByLink(ctx, link)
		if err != nil {
			log.Errorf("query file size by link error: %v, link: %s", err, link)
			size = 3 * util.GB
		}

		file, err := p.createFromLink(ctx, master, link, size, []string{})
		if err != nil {
			return nil, fmt.Errorf("create from link error: %v, link: %s", err, link)
		}

		linksStatusNotCompleted[link] = file
	}

	for _, f := range linksStatusNotCompleted {
		originalLink := hashToLink[f.OriginalLinkHash]
		sharedLinks[originalLink] = &share.Share{
			State:          share.StatusCreated,
			Title:          f.Name,
			HostSharedLink: "",
			OriginalLink:   originalLink,
			CreatedBy:      createBy,
			CreatedAt:      time.Now(),
			Size:           f.Size,
		}
	}
	if len(linksStatusNotCompleted) > 0 {
		log.WithContext(ctx).Debug("links not completed:", lo.Keys(linksStatusNotCompleted))
	}

	for _, f := range linksStatusOK {
		sharedLink, err := p.api.CreateShare(ctx, f.MasterUserID, f.WorkerUserID, f.FileID)
		if err != nil {
			return nil, err
		}
		originalLink := hashToLink[f.OriginalLinkHash]
		sharedLinks[originalLink] = &share.Share{
			State:          share.StatusFromFileStatus(f.Status),
			Title:          f.Name,
			HostSharedLink: sharedLink,
			OriginalLink:   originalLink,
			CreatedBy:      createBy,
			CreatedAt:      time.Now(),
			Size:           f.Size,
		}
	}

	for _, f := range linksStatusError {
		originalLink := hashToLink[f.OriginalLinkHash]
		sh := &share.Share{
			State:          share.StatusError,
			Title:          f.Name,
			HostSharedLink: "",
			OriginalLink:   originalLink,
			CreatedBy:      createBy,
			CreatedAt:      f.CreatedAt,
			Size:           f.Size,
			Error:          f.Error,
		}
		if isSensitiveLink(f.Error) {
			sh.State = share.StatusSensitive
		}
		sharedLinks[originalLink] = sh
	}

	return sharedLinks, nil
}

func isSensitiveLink(err string) bool {
	keywords := []string{"copyright", "harmful", "sensitive", "no longer available"}
	for _, v := range keywords {
		if strings.Contains(err, v) {
			return true
		}
	}
	return false
}

func (p *PikPak) createFromLink(ctx context.Context, master *model.MasterAccount, link string, size int64, excludeWorkers []string) (file *model.File, err error) {
	log := log.WithContext(ctx).WithFields(log.Fields{"tryFree": 1, "link": link})

	var worker *model.WorkerAccount
	workerAccountType := account.NotPremium
	if size >= 6*util.GB {
		workerAccountType = account.IsPremium
	}
	// try to get the limit_size is 0 byte account, if not exist, create one
	worker, err = p.m.GetWorkerWithEnoughCapacity(ctx, master.UserID, size, workerAccountType, excludeWorkers)
	if err != nil {
		return nil, fmt.Errorf("get worker err: %v, link: %s", err, link)
	}

	if worker == nil {
		return nil, fmt.Errorf("worker is nil, link: %s", link)
	}
	log.Debugf("use worker: %v, account type: %v for link: %s", worker, workerAccountType, link)
	file, err = p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if err != nil {
		invalidUntil, match := getInvalidUntilByCreateLinkError(err, workerAccountType)
		if match {
			err := p.m.UpdateAccountInvalidUtil(ctx, worker, invalidUntil)
			if err != nil {
				log.WithContext(ctx).WithField("worker", worker).Errorf("update account invalid util err: %v", err)
			}
		}

		if api.IsShouldNotRetryError(err) {
			log.WithContext(ctx).Errorf("create from link err: %v, should not retry", err)
			return nil, err
		}

		maybeSize := size
		if api.IsSpaceNotEnoughErr(err) {
			if workerAccountType == account.IsPremium {
				// if current use premium account, try to select other premium account with more space
				maybeSize = size + (2 * util.GB)
			} else {
				maybeSize = int64(6 * util.GB)
			}
		}
		go func() {
			// update the account premium expiration info
			_ = p.api.UpdateWorkerStorage(context.Background(), worker.UserID)
		}()
		excludeWorkers = append(excludeWorkers, worker.UserID)
		return p.createFromLink(ctx, master, link, maybeSize, excludeWorkers)
	}

	return file, nil
}

func getInvalidUntilByCreateLinkError(err error, workerAccountType account.Status) (invalidUntil time.Time, match bool) {
	if err == nil {
		return time.Now(), false
	}

	type UntilEntry struct {
		Fn   func(error) bool
		Time time.Time
	}
	forever := time.Hour * 24 * 365 * 100
	afterForever := time.Now().Add(forever)
	afterOneHour := time.Now().Add(time.Hour)
	afterOneDay := time.Now().Add(time.Hour * 24)
	// free account space less than 1GB, it will be invalid forever when the quota status update
	freeAccountInvalidUtils := []UntilEntry{
		{api.IsTaskRunNumsLimitErr, afterForever},
		{api.IsTaskDailyCreateLimitErr, afterForever},
	}
	// premium account space less than 6GB, it will be invalid forever when the quota status update
	premiumAccountInvalidUtils := []UntilEntry{
		{api.IsTaskRunNumsLimitErr, afterOneHour},
		{api.IsTaskDailyCreateLimitErr, afterOneDay},
	}

	invalidUtils := freeAccountInvalidUtils
	if workerAccountType == account.IsPremium {
		invalidUtils = premiumAccountInvalidUtils
	}

	if ue, ok := lo.Find(invalidUtils, func(k UntilEntry) bool {
		return k.Fn(err)
	}); ok {
		match = true
		invalidUntil = ue.Time
	}

	return
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// queryFilesizeByLink query file size by link (call whatslink.info api)
func (p *PikPak) queryFilesizeByLink(ctx context.Context, link string) (int64, error) {
	addr := fmt.Sprintf("https://whatslink.info/api/v1/link?url=%s", link)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bs, _ := io.ReadAll(resp.Body)
	var sizeInfo struct {
		Size int64 `json:"size"`
	}
	err = json.Unmarshal(bs, &sizeInfo)
	if err != nil {
		return 0, err
	}

	return sizeInfo.Size, nil
}

// GetLinkAccessInfos get link access log list
func GetLinkAccessInfos(ctx context.Context, originalLinkHash string) ([]string, error) {
	accessLog, err := config.Redis().Get(ctx, originalLinkHash).Result()
	if err != nil {
		accessLog = "[]"
	}

	var accessInfos []string
	err = json.Unmarshal([]byte(accessLog), &accessInfos)
	if err != nil {
		return nil, err
	}

	return accessInfos, err
}
