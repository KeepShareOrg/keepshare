package account

import (
	"context"
	"encoding/json"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/api"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/hibiken/asynq"
	"gorm.io/gen"
	"time"
)

func (m *Manager) keepTokenAliveWorker() {
	ctx := context.Background()
	ticker := time.NewTicker(time.Hour * 24)

	for {
		accounts, err := m.queryShouldKeepAliveMasterUserIds(ctx)
		if err == nil && len(accounts) > 0 {
			for _, uid := range accounts {
				log.Debugf("should login account user_id: %v", uid)
				_, err := m.api.CreateToken(ctx, uid, true)
				if err != nil {
					log.Errorf("keep alive, create token err: %v", err)
					// remove invalid password
					ma := m.q.MasterAccount
					ma.WithContext(ctx).Where(ma.UserID.Eq(uid)).Update(ma.Password, "")
				}
			}
		}

		_ = m.tokenTaskEnqueue(ctx)

		<-ticker.C
	}
}

func (m *Manager) queryShouldKeepAliveMasterUserIds(ctx context.Context) ([]string, error) {
	ma := m.q.MasterAccount
	tk := m.q.Token
	conditions := []gen.Condition{
		ma.KeepshareUserID.Neq(""),
		ma.Password.Neq(""),
		ma.Columns(ma.UserID).NotIn(tk.WithContext(ctx).Select(tk.UserID)),
	}
	var userIds []string
	err := ma.WithContext(ctx).Where(conditions...).Pluck(ma.UserID, &userIds)
	if err != nil {
		return nil, err
	}

	log.Debugf("should login account length: %v", len(userIds))
	return userIds, nil
}

const (
	taskTypeRefreshToken = "refresh_token"
)

type RefreshTokenTask struct {
	UserID       string `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}

func (m *Manager) registerRefreshTokenTask() {
	err := m.Queue.RegisterHandler(taskTypeRefreshToken, asynq.HandlerFunc(m.handleRefreshToken))
	if err != nil {
		log.Errorf("register refresh token task error: %v", err)
	}
}

func (m *Manager) handleRefreshToken(ctx context.Context, task *asynq.Task) error {
	var t RefreshTokenTask
	if err := json.Unmarshal(task.Payload(), &t); err != nil {
		return err
	}

	_, err := m.api.RefreshToken(ctx, t.UserID, t.RefreshToken)
	if err != nil {
		if api.IsInvalidGrantErr(err) {
			//if refresh token failed, delete the token
			m.q.Token.WithContext(ctx).Where(m.q.Token.RefreshToken.Eq(t.RefreshToken)).Delete()
			return nil
		}
		return err
	}

	return nil
}

func (m *Manager) tokenTaskEnqueue(ctx context.Context) error {
	tk := m.q.Token
	pageSize := 5000
	pageIndex := 0

	for {
		tokenInfos, err := tk.WithContext(ctx).
			Where(tk.CreatedAt.Lte(time.Now().Add(-10 * time.Minute))).
			Order(tk.CreatedAt).
			Offset(pageIndex * pageSize).
			Limit(pageSize).
			Find()
		if err != nil || len(tokenInfos) <= 0 {
			break
		}

		for _, info := range tokenInfos {
			payload, _ := json.Marshal(RefreshTokenTask{
				UserID:       info.UserID,
				RefreshToken: info.RefreshToken,
			})

			if _, err := m.Queue.Enqueue(taskTypeRefreshToken, payload, asynq.Queue(constant.AsyncQueueRefreshToken)); err != nil {
				log.WithField("user_id", info.UserID).Errorf("enqueue refresh token task error: %v", err)
			}
		}

		pageIndex++
	}

	return nil
}
