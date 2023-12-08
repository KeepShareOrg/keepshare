// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package account

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/model"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/mail"
	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
	"gorm.io/gen"
	"gorm.io/gorm"
)

const taskTypeInviteSubAccount = "invite_sub_account"

// Manager accounts.
type Manager struct {
	*hosts.Dependencies

	q   *query.Query
	api *api.API

	masterBufferSize        int
	masterBufferConcurrency int
	masterBufferInterval    time.Duration

	workerBufferSize        int
	workerBufferConcurrency int
	workerBufferInterval    time.Duration
}

// NewManager returns a manager instance.
func NewManager(q *query.Query, api *api.API, d *hosts.Dependencies) *Manager {
	m := &Manager{
		Dependencies: d,

		q:   q,
		api: api,

		masterBufferSize:        5,
		masterBufferConcurrency: 1,
		masterBufferInterval:    5 * time.Second,

		workerBufferSize:        10,
		workerBufferConcurrency: 2,
		workerBufferInterval:    1 * time.Second,
	}

	m.initConfig()

	m.Queue.RegisterHandler(taskTypeInviteSubAccount, asynq.HandlerFunc(m.inviteSubAccount))

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
	master, err := m.createMaster(ctx, keepShareUserID)
	if err != nil {
		return nil, fmt.Errorf("create master err: %w", err)
	}
	return master, nil
}

func (m *Manager) createMaster(ctx context.Context, keepShareUserID string) (*model.MasterAccount, error) {
	var ret gen.ResultInfo
	var master *model.MasterAccount
	now := time.Now().Round(time.Second)

	err := m.q.Transaction(func(tx *query.Query) error {
		t := &tx.MasterAccount
		var err error
		ret, err = t.WithContext(ctx).
			Where(t.KeepshareUserID.Eq("")).
			Select(t.KeepshareUserID, t.UpdatedAt).
			Order(t.CreatedAt).
			Limit(1).
			Updates(&model.MasterAccount{
				KeepshareUserID: keepShareUserID,
				UpdatedAt:       now,
			})
		if err != nil {
			return fmt.Errorf("bind new master err: %w", err)
		}
		if ret.RowsAffected < 1 {
			return nil
		}
		master, err = t.WithContext(ctx).Where(t.KeepshareUserID.Eq(keepShareUserID), t.UpdatedAt.Eq(now)).Take()
		if err != nil {
			return err
		}

		if err = m.api.JoinReferral(ctx, master.UserID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if master == nil {
		return nil, errors.New("no enough master account")
	}
	return master, nil
}

// GetWorkerWithEnoughCapacity get or create a worker account with sufficient capacity
// and with the smallest remaining capacity.
func (m *Manager) GetWorkerWithEnoughCapacity(ctx context.Context, master string, size int64, status Status, excludeWorkers []string) (*model.WorkerAccount, error) {
	a, err := m.getWorkerWithEnoughCapacity(ctx, master, size, status, excludeWorkers)
	if err != nil && !gormutil.IsNotFoundError(err) {
		return nil, fmt.Errorf("get worker err: %w", err)
	}

	if a == nil {
		a, err = m.CreateWorker(ctx, master, status)
		if err != nil {
			log.Debugf("create worker err: %v", err)
			// if create worker failed, try to use premium worker
			a, err := m.getWorkerWithEnoughCapacity(ctx, master, size, IsPremium, excludeWorkers)
			if err != nil {
				return nil, err
			}
			return a, nil
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
func (m *Manager) getWorkerWithEnoughCapacity(ctx context.Context, master string, size int64, status Status, excludeWorkers []string) (*model.WorkerAccount, error) {
	t := &m.q.WorkerAccount
	f := &m.q.File
	// query the worker account that running task less 100
	if len(excludeWorkers) == 0 {
		excludeWorkers = []string{""}
	}
	ews := make([]string, len(excludeWorkers))
	for i, v := range excludeWorkers {
		ews[i] = fmt.Sprintf("'%s'", v)
	}
	ids := strings.Join(ews, ",")
	var res []*model.WorkerAccount
	w := fmt.Sprintf(`select * from %s pwa inner join(select distinct pwa.user_id from %s as pf left join %s as pwa on pf.worker_user_id = pwa.user_id where pwa.master_user_id = '%s' and pf.status = '%s' and pwa.user_id not in (%s) group by pwa.user_id having count(*) < 100) as sub_pwa on pwa.user_id = sub_pwa.user_id and pwa.invalid_until <= now() and pwa.limit_size > pwa.used_size + %v limit 1`,
		t.TableName(), f.TableName(), t.TableName(), master, comm.StatusRunning, ids, size)
	if err := config.MySQL().Raw(w).Scan(&res).Error; err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return res[0], nil
}

type inviteSubAccountRequest struct {
	MasterUserID string
	WorkerUserID string
	WorkerEmail  string
}

func (m *Manager) CreateWorker(ctx context.Context, master string, status Status) (*model.WorkerAccount, error) {
	worker, err := m.createWorker(ctx, master, status)
	if err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(&inviteSubAccountRequest{
		MasterUserID: worker.MasterUserID,
		WorkerUserID: worker.UserID,
		WorkerEmail:  worker.Email,
	})
	_, _ = m.Queue.Enqueue(taskTypeInviteSubAccount, payload, asynq.MaxRetry(3))

	return worker, nil
}

// CreateWorker creates a worker for the master.
func (m *Manager) createWorker(ctx context.Context, master string, status Status) (*model.WorkerAccount, error) {
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
		Order(t.CreatedAt).
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
			return m.createWorker(ctx, master, IsPremium)
		}
		return nil, errors.New("no enough worker account")
	}

	return t.WithContext(ctx).Where(t.MasterUserID.Eq(master), t.UpdatedAt.Eq(now)).Take()
}

// UpdateAccountInvalidUtil update worker invalid until
func (m *Manager) UpdateAccountInvalidUtil(ctx context.Context, worker *model.WorkerAccount, until time.Time) error {
	if _, err := m.q.WorkerAccount.WithContext(ctx).
		Where(m.q.WorkerAccount.UserID.Eq(worker.UserID)).
		Updates(&model.WorkerAccount{
			InvalidUntil: until,
		}); err != nil {
		return fmt.Errorf("update worker err: %w", err)
	}
	return nil
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

func (m *Manager) inviteSubAccount(ctx context.Context, task *asynq.Task) (err error) {
	var req inviteSubAccountRequest
	_ = json.Unmarshal(task.Payload(), &req)
	if req.MasterUserID == "" || req.WorkerEmail == "" {
		log.Debugf("task: %s, invalid msg: %s", task.Type(), task.Payload())
		return nil
	}

	l := log.WithFields(log.Fields{
		"master": req.MasterUserID,
		"worker": req.WorkerUserID,
		"email":  req.WorkerEmail,
	})
	defer func() {
		if err != nil {
			l.WithError(err).Error("inviteSubAccount err")
		} else {
			l.Debug("inviteSubAccount ok")
		}
	}()

	// send invite email
	sendTime := time.Now()
	err = m.api.InviteSubAccount(ctx, req.MasterUserID, req.WorkerEmail)
	if err != nil {
		return fmt.Errorf("send invite request err: %w", err)
	}

	// verify email
	var verifyURL string
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		verifyURL, _, err = m.getInviteURL(ctx, req.WorkerEmail, sendTime)
		if verifyURL != "" {
			break
		}
	}
	if verifyURL == "" {
		return fmt.Errorf("invite sub account email not found err: %v", err)
	}

	u, err := url.Parse(verifyURL)
	if err != nil || u.RawQuery == "" {
		return fmt.Errorf("invalid verify url: %s", verifyURL)
	}

	token := u.Query().Get("token")
	if len(token) < 10 {
		return fmt.Errorf("invalid verify url: %s", verifyURL)
	}

	err = m.api.VerifyInviteSubAccountToken(ctx, token)
	if err != nil {
		return fmt.Errorf("verify invite url err: %w", err)
	}
	return nil
}

var (
	inviteURLRegexp     = regexp.MustCompile(`https://mypikpak.com/referral/verify\?[^\s'")]*`)
	inviteURLFromRegexp = regexp.MustCompile(`support@mailer.mypikpak.com`)
	inviteURLHTTPClient = &http.Client{Timeout: 5 * time.Second}
)

func (m *Manager) getInviteURL(ctx context.Context, email string, sentTime time.Time) (url string, found bool, err error) {
	url, found, err = mail.FindText(ctx, m.Mailer, email, inviteURLRegexp, &mail.Filter{
		SendTime:   sentTime,
		FromRegexp: inviteURLFromRegexp,
	})
	if err != nil {
		log.WithField("email", email).WithError(err).Error("getInviteURL err")
	} else {
		log.WithField("email", email).Debugf("getInviteURL found: %t, url: %s", found, url)
	}
	return
}
