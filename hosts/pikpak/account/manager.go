// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	log "github.com/sirupsen/logrus"
	"gorm.io/gen"
)

// Manager accounts.
type Manager struct {
	*hosts.Dependencies

	q   *query.Query
	api *api.API

	masterBufferSize     int
	masterBufferInterval time.Duration

	workerBufferSize     int
	workerBufferInterval time.Duration
}

// NewManager returns a manager instance.
func NewManager(q *query.Query, api *api.API, d *hosts.Dependencies) *Manager {
	m := &Manager{
		Dependencies: d,

		q:   q,
		api: api,

		masterBufferSize:     2,
		masterBufferInterval: 10 * time.Second,
		workerBufferSize:     5,
		workerBufferInterval: 1 * time.Second,
	}

	m.initConfig()

	go m.checkMasterBuffer()
	go m.checkWorkerBuffer()

	return m
}

// GetMaster get bound account, if not found, create a master account and bind to this user.
func (m *Manager) GetMaster(ctx context.Context, keepShareUserID string) (*model.MasterAccount, error) {
	if keepShareUserID == "" {
		return nil, errors.New("get master with invalid user id")
	}

	a, err := m.q.MasterAccount.WithContext(ctx).Where(m.q.MasterAccount.KeepshareUserID.Eq(keepShareUserID)).Take()
	if err != nil && !gormutil.IsNotFoundError(err) {
		return nil, fmt.Errorf("get master err: %w", err)
	}

	if a != nil {
		return a, nil
	}

	return m.createMaster(ctx, keepShareUserID)
}

func (m *Manager) createMaster(ctx context.Context, keepShareUserID string) (*model.MasterAccount, error) {
	t := &m.q.MasterAccount

	now := time.Now().Round(time.Second)
	ret, err := t.WithContext(ctx).
		Where(t.KeepshareUserID.Eq("")).
		Select(t.KeepshareUserID, t.UpdatedAt).
		Limit(1).
		Updates(&model.MasterAccount{
			KeepshareUserID: keepShareUserID,
			UpdatedAt:       now,
		})
	if err != nil {
		return nil, fmt.Errorf("bind new master err: %w", err)
	}
	if ret.RowsAffected < 1 {
		return nil, errors.New("no enough worker account")
	}

	return t.WithContext(ctx).Where(t.KeepshareUserID.Eq(keepShareUserID), t.UpdatedAt.Eq(now)).Take()
}

// GetWorkerWithEnoughCapacity get or create a worker account with sufficient capacity
// and with the smallest remaining capacity.
func (m *Manager) GetWorkerWithEnoughCapacity(ctx context.Context, master string, size int64, status Status) (*model.WorkerAccount, error) {
	a, err := m.getWorkerWithEnoughCapacity(ctx, master, size, status)
	if err != nil && !gormutil.IsNotFoundError(err) {
		return nil, fmt.Errorf("get worker err: %w", err)
	}

	if a == nil {
		a, err = m.CreateWorker(ctx, master, status)
		if err != nil {
			return nil, err
		}
	}

	if err := m.api.UpdateWorkerPremium(ctx, a); err != nil {
		log.Errorf("update worker err: %v", err)
	}

	return a, nil
}

// Status identifies whether this account is a premium.
type Status int

// enum statuses.
const (
	whatever Status = iota
	IsPremium
	NotPremium
)

func (f Status) where(q *query.Query) gen.Condition {
	switch f {
	case IsPremium:
		return q.WorkerAccount.PremiumExpiration.Gte(time.Now())
	case NotPremium:
		return q.WorkerAccount.PremiumExpiration.Lt(time.Now())
	case whatever:
		return nil
	default:
		return nil
	}
}

// getWorkerWithEnoughCapacity returns the worker with enough and max free size
func (m *Manager) getWorkerWithEnoughCapacity(ctx context.Context, master string, size int64, status Status) (*model.WorkerAccount, error) {
	t := &m.q.WorkerAccount
	where := []gen.Condition{
		t.MasterUserID.Eq(master),
		t.LimitSize.GtCol(t.UsedSize.Add(size)),
	}
	if w := status.where(m.q); w != nil {
		where = append(where, w)
	}

	return t.WithContext(ctx).Where(where...).Order(t.UsedSize.SubCol(t.LimitSize)).Limit(1).Take()
}

// CreateWorker creates a worker for the master.
func (m *Manager) CreateWorker(ctx context.Context, master string, status Status) (*model.WorkerAccount, error) {
	t := &m.q.WorkerAccount

	if status == IsPremium {
		count, err := t.Where(t.MasterUserID.Eq(master), t.PremiumExpiration.Gt(time.Now())).Count()
		if err != nil {
			return nil, fmt.Errorf("count workers err: %w", err)
		}
		if count >= comm.MaxPremiumWorkers {
			return nil, fmt.Errorf("current number of premium workers is %d, reached the limit: %d", count, comm.MaxPremiumWorkers)
		}
	}

	where := []gen.Condition{
		t.MasterUserID.Eq(""),
	}

	tmpFlag := status
	if tmpFlag == whatever {
		tmpFlag = NotPremium
	}
	if w := tmpFlag.where(m.q); w != nil {
		where = append(where, w)
	}

	now := time.Now().Round(time.Second)
	ret, err := t.WithContext(ctx).
		Where(where...).
		Select(t.MasterUserID, t.UpdatedAt).
		Limit(1).
		Updates(&model.WorkerAccount{
			MasterUserID: master,
			UpdatedAt:    now,
		})
	if err != nil {
		return nil, fmt.Errorf("bind new worker err: %w", err)
	}
	if ret.RowsAffected < 1 {
		if status == whatever {
			// try to get a premium account.
			return m.CreateWorker(ctx, master, IsPremium)
		}
		return nil, errors.New("no enough worker account")
	}

	return t.WithContext(ctx).Where(t.MasterUserID.Eq(master), t.UpdatedAt.Eq(now)).Take()
}

type (
	// CountWorkersResponse is the result.
	CountWorkersResponse struct {
		Premium countWorkersData `json:"premium"`
		Free    countWorkersData `json:"free"`
	}
	countWorkersData struct {
		Premium int `json:"-"`
		Count   int `json:"count"`
		Used    int `json:"used"`
		Limit   int `json:"limit"`
	}
)

// CountWorkers counts total workers info for the master.
func (m *Manager) CountWorkers(ctx context.Context, master string) (*CountWorkersResponse, error) {
	t := &m.q.WorkerAccount
	sel := fmt.Sprintf(
		"IF(`%s` > NOW(), 1, 0) AS `premium`, COUNT(*) AS count, SUM(`%s`) AS used, SUM(`%s`) AS `limit`",
		t.PremiumExpiration.ColumnName().String(),
		t.UsedSize.ColumnName().String(),
		t.LimitSize.ColumnName().String(),
	)

	var ret []countWorkersData
	err := m.Mysql.WithContext(ctx).
		Table(model.TableNameWorkerAccount).
		Select(sel).
		Where(fmt.Sprintf("`%s` = ?", t.MasterUserID.ColumnName().String()), master).
		Group("`premium`").
		Find(&ret).Error
	if err != nil && !gormutil.IsNotFoundError(err) {
		log.WithError(err).Error("count workers err")
		return nil, err
	}

	var resp CountWorkersResponse
	for _, data := range ret {
		if data.Premium == 1 {
			resp.Premium = data
		} else {
			resp.Free = data
		}
	}
	return &resp, nil
}
