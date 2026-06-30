// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package forward

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	pikpakquery "github.com/KeepShareOrg/keepshare/hosts/pikpak/query"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/mail"
	serverquery "github.com/KeepShareOrg/keepshare/server/query"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// SendFunc is the hook used by Run to deliver a built message. Tests inject a
// recorder; production uses Sender.Send.
type SendFunc func(b *Built, acc config.ForwardAccount) error

// loadMastersFunc returns the set of currently-registered pikpak master emails.
type loadMastersFunc func(ctx context.Context) (map[string]struct{}, error)

// lookupUserFunc returns the keepshare user email for an original recipient.
type lookupUserFunc func(ctx context.Context, originalTo string) (string, error)

// fetchBodyFunc fetches the plain-text body of a message from the mail server.
// The WebSocket notification frame carries only headers, so the body must be
// pulled separately by mailbox + message id.
type fetchBodyFunc func(ctx context.Context, mailbox, id string) (string, error)

// newMonitorFn is overridable from tests. Production builds a *Monitor.
var newMonitorFn = func(wsURL string) (MessageSource, error) {
	if wsURL == "" {
		return nil, errors.New("forward: monitor URL is empty")
	}
	return NewMonitor(wsURL)
}

// Run blocks until ctx is cancelled, forwarding matching messages. It returns
// nil on clean shutdown. When the feature is disabled or misconfigured it
// logs and returns nil so the caller's goroutine exits quietly.
func Run(ctx context.Context, db *gorm.DB, rdb *redis.Client, mailer mail.Mailer) error {
	cfg := config.ForwardMails()
	if !cfg.Enable {
		return nil
	}
	if len(cfg.Accounts) == 0 {
		log.Warn("[forward] enabled but no accounts configured; exiting")
		return nil
	}
	if mailer == nil {
		log.Warn("[forward] enabled but mailer is nil; exiting")
		return nil
	}

	wsURL, err := resolveMonitorURL(cfg.MonitorURL, viper.GetString("mail_server"))
	if err != nil {
		log.Warnf("[forward] no monitor URL resolved: %v", err)
		return nil
	}
	src, err := newMonitorFn(wsURL)
	if err != nil {
		return fmt.Errorf("forward: monitor: %w", err)
	}

	rotator := NewRotator(cfg.Accounts, 5*time.Minute)
	dedup := NewDeduper(rdb, 24*time.Hour)
	sender := NewSender(cfg.Notice)

	loadMasters := func(ctx context.Context) (map[string]struct{}, error) {
		return loadMasterEmails(ctx, db)
	}
	lookupUser := func(ctx context.Context, originalTo string) (string, error) {
		return lookupUserEmail(ctx, db, originalTo)
	}
	fetchBody := func(ctx context.Context, mailbox, id string) (string, error) {
		b, err := mailer.Get(ctx, mailbox, id)
		if err != nil {
			return "", err
		}
		return b.Text, nil
	}
	sendFn := func(b *Built, acc config.ForwardAccount) error {
		return sender.Send(b, acc)
	}

	log.WithFields(log.Fields{
		"monitor":   wsURL,
		"accounts":  len(cfg.Accounts),
		"dedup_ttl": "24h",
	}).Info("[forward] enabled")

	return runWith(ctx, rotator, dedup, sender, src, loadMasters, lookupUser, fetchBody, sendFn)
}

// runWith is the orchestration loop. All dependencies are passed explicitly so
// tests can substitute fakes for the message source, loaders and sender.
func runWith(
	ctx context.Context,
	rotator *Rotator,
	dedup *Deduper,
	sender *Sender,
	src MessageSource,
	loadMasters loadMastersFunc,
	lookupUser lookupUserFunc,
	fetchBody fetchBodyFunc,
	send SendFunc,
) error {
	masters, err := loadMasters(ctx)
	if err != nil {
		return fmt.Errorf("forward: load masters: %w", err)
	}
	refresh := time.NewTicker(5 * time.Minute)
	defer refresh.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-refresh.C:
			if mp, err := loadMasters(ctx); err == nil {
				masters = mp
			}
		default:
		}

		msg, ok, err := src.Next(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			log.Debugf("[forward] source next err: %v", err)
			continue
		}
		if !ok {
			continue
		}

		matched, reason, accepted := Filter(msg, masters)
		if !accepted {
			log.Debugf("[forward] drop: %s (msg=%s)", reason, msg.ID)
			continue
		}

		fp := msg.Fingerprint()
		first, err := dedup.FirstSeen(ctx, fp)
		if err != nil {
			log.Warnf("[forward] dedup err: %v", err)
		}
		if !first {
			log.Debugf("[forward] drop: already seen (msg=%s)", msg.ID)
			continue
		}

		toUser, err := lookupUser(ctx, matched)
		if err != nil {
			log.Warnf("[forward] cannot resolve user for %s: %v", maskEmail(matched), err)
			_ = dedup.Forget(ctx, fp)
			continue
		}

		// The notification frame carries only headers; fetch the body now.
		body, err := fetchBody(ctx, msg.Mailbox, msg.ID)
		if err != nil {
			log.Warnf("[forward] fetch body err (mailbox=%s id=%s): %v", maskEmail(msg.Mailbox), msg.ID, err)
			_ = dedup.Forget(ctx, fp)
			continue
		}

		acc, more := rotator.Next()
		if !more {
			log.Warnf("[forward] all accounts cooling down; drop msg=%s", msg.ID)
			_ = dedup.Forget(ctx, fp)
			continue
		}

		b, err := sender.Build(acc, toUser, matched, msg, body)
		if err != nil {
			log.Warnf("[forward] build err: %v", err)
			_ = dedup.Forget(ctx, fp)
			continue
		}
		if err := send(b, acc); err != nil {
			log.Warnf("[forward] send err via %s: %v (cooldown 5m)", acc.Email, err)
			rotator.MarkFailed(acc.Email)
			_ = dedup.Forget(ctx, fp)
			continue
		}
		rotator.MarkOK(acc.Email)
		log.Infof("[forward] ok msg=%s to=%s via=%s", shortID(msg.ID), maskEmail(toUser), acc.Email)
	}
}

func shortID(s string) string {
	if len(s) <= 8 {
		return s
	}
	return s[:8]
}

// maskEmail hides the local part of an email for logs, keeping just enough to
// correlate without leaking the full address.
func maskEmail(email string) string {
	at := -1
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			at = i
			break
		}
	}
	if at <= 0 {
		return "***"
	}
	local := email[:at]
	if len(local) <= 1 {
		return "*" + email[at:]
	}
	return local[:1] + "***" + email[at:]
}

// loadMasterEmails returns the set of currently-registered pikpak master emails.
func loadMasterEmails(ctx context.Context, db *gorm.DB) (map[string]struct{}, error) {
	q := pikpakquery.Use(db).MasterAccount
	rows, err := q.WithContext(ctx).Select(q.Email).Find()
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if r.Email != "" {
			out[r.Email] = struct{}{}
		}
	}
	return out, nil
}

// lookupUserEmail returns the keepshare_user.email for the master account
// addressed by originalTo, via the master → keepshare_user_id → user hop.
func lookupUserEmail(ctx context.Context, db *gorm.DB, originalTo string) (string, error) {
	mq := pikpakquery.Use(db).MasterAccount
	ma, err := mq.WithContext(ctx).Where(mq.Email.Eq(originalTo)).Take()
	if err != nil {
		return "", err
	}
	if ma.KeepshareUserID == "" {
		return "", fmt.Errorf("master account %s not bound", originalTo)
	}
	uq := serverquery.Use(db).User
	u, err := uq.WithContext(ctx).Where(uq.ID.Eq(ma.KeepshareUserID)).Take()
	if err != nil {
		return "", err
	}
	if u.Email == "" {
		return "", fmt.Errorf("keepshare user %s has no email", ma.KeepshareUserID)
	}
	return u.Email, nil
}
