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
	"github.com/KeepShareOrg/keepshare/pkg/share"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

// CreateFromLinks create shared links based on the input original links.
func (p *PikPak) CreateFromLinks(ctx context.Context, keepShareUserID string, originalLinks []string, createBy string) (sharedLinks map[string]*share.Share, err error) {
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

	files, err := p.q.File.WithContext(ctx).Where(p.q.File.OriginalLinkHash.In(hashes...), p.q.File.MasterUserID.Eq(master.UserID)).Find()
	if err != nil && !gormutil.IsNotFoundError(err) {
		return nil, fmt.Errorf("query files err: %w", err)
	}

	linksCompleted := map[string]*model.File{}
	linksPending := map[string]*model.File{}
	for _, f := range files {
		l := hashToLink[f.OriginalLinkHash]
		switch f.Status {
		case comm.StatusOK:
			linksCompleted[l] = f
		case comm.StatusError:
			// delete error files
			p.q.File.WithContext(ctx).Delete(f)
		default:
			// TODO delete timeout files
			linksPending[l] = f
		}
	}

	// only create files for new links.
	for _, link := range originalLinks {
		if linksCompleted[link] != nil || linksPending[link] != nil {
			continue
		}

		file, err := p.createFromLink(ctx, master, link)
		if err != nil {
			return nil, err
		}

		linksPending[link] = file
	}

	//const intervalSeconds = 1
	//loop check if the task is done within N seconds.
	//for i := 0; i < comm.CreateFileSyncWaitSeconds/intervalSeconds && len(linksPending) > 0; i++ {
	//	time.Sleep(intervalSeconds * time.Second)
	//	workerFiles := map[string][]*model.File{}
	//	for _, file := range linksPending {
	//		workerFiles[file.WorkerUserID] = append(workerFiles[file.WorkerUserID], file)
	//	}
	//	for worker, files := range workerFiles {
	//		err := p.api.UpdateFilesStatus(ctx, worker, files)
	//		if err != nil {
	//			continue
	//		}
	//		for _, file := range files {
	//			if comm.IsFinalStatus(file.Status) {
	//				link := hashToLink[file.OriginalLinkHash]
	//				linksCompleted[link] = file
	//				delete(linksPending, link)
	//			}
	//		}
	//	}
	//}

	sharedLinks = map[string]*share.Share{}

	for _, f := range linksPending {
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
	if len(linksPending) > 0 {
		// TODO
		log.Debug("links not completed:", lo.Keys(linksPending))
	}

	for _, f := range linksCompleted {
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
	// firstly, try with a existed free worker and free size more than 1GB
	worker, err := p.m.GetWorkerWithEnoughCapacity(ctx, master.UserID, util.GB, account.NotPremium)
	if err != nil {
		return nil, err
	}
	file, err := p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if api.IsAccountLimited(err) {
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
	worker, err = p.m.CreateWorker(ctx, master.UserID, account.NotPremium)
	if err != nil {
		return nil, err
	}
	file, err = p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if api.IsAccountLimited(err) {
		goto tryWithPremiumAccount
	}
	if err != nil {
		return nil, err
	}
	return file, nil

tryWithPremiumAccount:
	worker, err = p.m.GetWorkerWithEnoughCapacity(ctx, master.UserID, 6*util.GB, account.IsPremium)
	if err != nil {
		return nil, err
	}
	file, err = p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if api.IsAccountLimited(err) {
		goto tryWithNewPremiumAccount
	}
	if err != nil {
		return nil, err
	}
	return file, nil

tryWithNewPremiumAccount:
	worker, err = p.m.CreateWorker(ctx, master.UserID, account.IsPremium)
	if err != nil {
		return nil, err
	}
	file, err = p.api.CreateFilesFromLink(ctx, master.UserID, worker.UserID, link)
	if err != nil {
		return nil, err
	}
	return file, nil
}
