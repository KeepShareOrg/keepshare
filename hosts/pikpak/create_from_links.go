// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package pikpak

import (
	"context"
	"fmt"
	"time"

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

// CreateFromLinks create shared links based on the input original links.
func (p *PikPak) CreateFromLinks(ctx context.Context, keepShareUserID string, originalLinks []string, createBy string) (sharedLinks map[string]*share.Share, err error) {
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
	linksStatusPending := map[string]*model.File{}
	for _, f := range files {
		l := hashToLink[f.OriginalLinkHash]
		switch f.Status {
		case comm.StatusOK:
			linksStatusOK[l] = f
		case comm.StatusError:
			// delete error files
			p.q.File.WithContext(ctx).Delete(f)
		default:
			// TODO delete timeout files
			linksStatusPending[l] = f
			p.api.TriggerRunningFile(f)
		}
	}

	// only create files for new links.
	for _, link := range originalLinks {
		if linksStatusOK[link] != nil || linksStatusPending[link] != nil {
			continue
		}

		file, err := p.createFromLink(ctx, master, link)
		if err != nil {
			return nil, fmt.Errorf("creat from link error: %v", err)
		}

		linksStatusPending[link] = file
	}

	sharedLinks = map[string]*share.Share{}
	for _, f := range linksStatusPending {
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
	if len(linksStatusPending) > 0 && log.IsDebugEnabled() {
		log.WithContext(ctx).Debug("links not completed:", lo.Keys(linksStatusPending))
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

	return sharedLinks, nil
}

func (p *PikPak) createFromLink(ctx context.Context, master *model.MasterAccount, link string) (*model.File, error) {
	var excludeWorkers []string
	var tryPremium int

	log.ContextWithFields(ctx, log.Fields{"tryFree": 1})
	// firstly, try with an existed free worker and free size more than 1GB
	worker, err := p.m.GetWorkerWithEnoughCapacity(ctx, master.UserID, util.GB, account.NotPremium, excludeWorkers)
	if err != nil {
		return nil, err
	}
	file, err := p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if api.IsAccountLimited(err) {
		invalidUtil := time.Now()
		if api.IsTaskDailyCreateLimitErr(err) {
			invalidUtil = time.Now().Add(24 * time.Hour)
		}
		if api.IsTaskRunNumsLimitErr(err) || api.IsSpaceNotEnoughErr(err) {
			invalidUtil = time.Now().Add(time.Hour)
		}
		if invalidUtil.Sub(time.Now()) > 0 {
			if err := p.m.UpdateAccountInvalidUtil(ctx, worker, invalidUtil); err != nil {
				log.WithContext(ctx).WithField("worker", worker).Errorf("update account invalid util err: %v", err)
			}
		}

		excludeWorkers = append(excludeWorkers, worker.UserID)
		if worker.LimitSize <= 0 || worker.LimitSize-worker.UsedSize > 5*util.GB {
			goto tryWithPremiumAccount
		} else {
			goto tryWithNewFreeAccount
		}
	}
	if err != nil {
		return nil, err
	}
	return file, nil

tryWithNewFreeAccount:
	log.ContextWithFields(ctx, log.Fields{"tryNewFree": 1})
	worker, err = p.m.CreateWorker(ctx, master.UserID, account.NotPremium)
	if err != nil {
		return nil, err
	}
	file, err = p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if api.IsAccountLimited(err) {
		invalidUtil := time.Now()
		if api.IsTaskDailyCreateLimitErr(err) {
			invalidUtil = time.Now().Add(24 * time.Hour)
		}
		if api.IsTaskRunNumsLimitErr(err) || api.IsSpaceNotEnoughErr(err) {
			invalidUtil = time.Now().Add(time.Hour)
		}
		if invalidUtil.Sub(time.Now()) > 0 {
			if err := p.m.UpdateAccountInvalidUtil(ctx, worker, invalidUtil); err != nil {
				log.WithContext(ctx).WithField("worker", worker).Errorf("update account invalid util err: %v", err)
			}
		}

		excludeWorkers = append(excludeWorkers, worker.UserID)
		goto tryWithPremiumAccount
	}
	if err != nil {
		return nil, err
	}
	return file, nil

tryWithPremiumAccount:
	tryPremium++
	log.ContextWithFields(ctx, log.Fields{"tryPremium": tryPremium})
	worker, err = p.m.GetWorkerWithEnoughCapacity(ctx, master.UserID, 6*util.GB, account.IsPremium, excludeWorkers)
	if err != nil {
		return nil, err
	}
	file, err = p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if api.IsAccountLimited(err) {
		invalidUtil := time.Now()
		if api.IsTaskDailyCreateLimitErr(err) {
			invalidUtil = time.Now().Add(24 * time.Hour)
		}
		if api.IsTaskRunNumsLimitErr(err) || api.IsSpaceNotEnoughErr(err) {
			invalidUtil = time.Now().Add(time.Hour)
		}
		if invalidUtil.Sub(time.Now()) > 0 {
			if err := p.m.UpdateAccountInvalidUtil(ctx, worker, invalidUtil); err != nil {
				log.WithContext(ctx).WithField("worker", worker).Errorf("update account invalid util err: %v", err)
			}
		}

		excludeWorkers = append(excludeWorkers, worker.UserID)
		goto tryWithPremiumAccount
	}
	if err != nil {
		return nil, err
	}
	return file, nil
}
