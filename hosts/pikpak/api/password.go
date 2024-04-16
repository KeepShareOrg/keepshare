package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/pkg/mail"
	"github.com/KeepShareOrg/keepshare/pkg/util"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"regexp"
	"time"
)

const (
	resetStatusTodo  = "TODO"
	resetStatusDone  = "DONE"
	resetStatusError = "Error"

	taskTypeResetPassword = "reset_password"
)

func (api *API) ResetPassword(ctx context.Context, email, password string, rememberMe bool) (string, error) {
	l := log.WithContext(ctx).WithField("email", email)

	var err error
	defer func() {
		if err != nil {
			l.WithError(err)
		}
	}()

	captchaToken, err := api.resetPasswordCaptcha(ctx, email)
	if err != nil {
		return "", err
	}

	sendTime := time.Now()
	verificationID, err := api.resetPasswordSendEmail(ctx, email, captchaToken)
	if err != nil {
		return "", err
	}

	taskID := fmt.Sprintf("%v", time.Now().Nanosecond())
	uuid, err := uuid.NewUUID()
	if err == nil {
		taskID = uuid.String()
	}

	api.Redis.SetEx(ctx, taskID, resetStatusTodo, time.Hour)

	payload, _ := json.Marshal(ResetPasswordTask{
		TaskID:         taskID,
		Email:          email,
		Password:       password,
		RememberMe:     rememberMe,
		VerificationID: verificationID,
		EmailSendTime:  sendTime,
	})
	_, err = api.Queue.Enqueue(
		taskTypeResetPassword,
		payload,
		asynq.Queue(constant.AsyncQueueResetPassword),
		asynq.ProcessAt(time.Now().Add(time.Second*5)),
	)
	if err != nil {
		log.Errorf("enqueue task error: %v", err)
		return "", fmt.Errorf("reset password enqueue task error: %v", err)
	}

	return taskID, nil
}

func (api *API) resetPasswordCaptcha(ctx context.Context, email string) (string, error) {
	var r struct {
		*RespErr
		CaptchaToken string `json:"captcha_token"`
	}

	b := util.ToJSON(map[string]any{
		"action":    "POST:/v1/auth/verification",
		"client_id": webClientID,
		"device_id": deviceID,
		"meta": map[string]string{
			"email": email,
		},
	})
	resp, err := resCli.R().SetContext(ctx).SetError(&r).SetResult(&r).SetBody(b).SetHeaders(map[string]string{
		"x-client-id": webClientID,
		"x-device-id": deviceID,
	}).Post(userURL("/v1/shield/captcha/init"))
	if err != nil {
		return "", err
	}

	log.WithContext(ctx).Debugf("get captcha token resp body: %s", resp.String())
	if err = r.Error(); err != nil {
		return "", err
	}

	return r.CaptchaToken, nil
}

func (api *API) resetPasswordSendEmail(ctx context.Context, email, captcha string) (string, error) {
	var r struct {
		*RespErr
		VerificationID string `json:"verification_id"`
	}

	b := util.ToJSON(map[string]any{
		"target":           "USER",
		"usage":            "PASSWORD_RESET",
		"locale":           "en-US",
		"email":            email,
		"selected_channel": 2,
		"client_id":        webClientID,
	})

	resp, err := resCli.R().SetContext(ctx).SetError(&r).SetResult(&r).
		SetHeaders(map[string]string{
			"x-captcha-token": captcha,
			"x-client-id":     webClientID,
			"x-device-id":     deviceID,
		}).SetBody(b).Post(userURL("/v1/auth/verification"))
	if err != nil {
		return "", err
	}

	log.WithContext(ctx).Debugf("get verification id resp body: %s", resp.String())

	if err = r.Error(); err != nil {
		return "", err
	}

	return r.VerificationID, nil
}

func (api *API) verifyResetCode(ctx context.Context, verificationID, code string) (string, error) {
	var r struct {
		*RespErr
		VerificationToken string `json:"verification_token"`
	}

	b := util.ToJSON(map[string]any{
		"verification_code": code,
		"verification_id":   verificationID,
		"client_id":         webClientID,
	})

	resp, err := resCli.R().SetContext(ctx).SetError(&r).SetResult(&r).SetBody(b).Post(userURL("/v1/auth/verification/verify"))
	if err != nil {
		return "", err
	}

	log.WithContext(ctx).Debugf("verify reset code resp body: %s", resp.String())

	if err = r.Error(); err != nil {
		return "", err
	}

	return r.VerificationToken, nil
}

func (api *API) resetPassword(ctx context.Context, email, verificationToken, password string) error {
	var r struct {
		*RespErr
	}

	b := util.ToJSON(map[string]any{
		"new_password":       password,
		"verification_token": verificationToken,
		"email":              email,
		"client_id":          webClientID,
	})

	resp, err := resCli.R().SetContext(ctx).SetError(&r).SetResult(&r).SetBody(b).Post(userURL("/v1/auth/reset"))
	if err != nil {
		return err
	}

	log.WithContext(ctx).Debugf("reset password resp body: %s", resp.String())

	if err = r.Error(); err != nil {
		return err
	}

	return nil
}

var (
	resetPasswordCodeRegexp      = regexp.MustCompile(`[0-9]{6}`)
	resetPasswordEmailFromRegexp = regexp.MustCompile(`noreply@accounts.mypikpak.com`)
)

func (api *API) resetPasswordGetCode(ctx context.Context, email string, sentTime time.Time) (code string, found bool, err error) {
	l := log.WithContext(ctx).WithField("email", email)
	code, found, err = mail.FindText(ctx, api.Mailer, email, resetPasswordCodeRegexp, &mail.Filter{
		SendTime:   sentTime,
		FromRegexp: resetPasswordEmailFromRegexp,
	})
	if err != nil {
		l.WithError(err).Error("resetPasswordGetCode err")
	} else {
		l.Infof("resetPasswordGetCode found: %t, code: %s", found, code)
	}
	return
}

func (api *API) RegisterResetPasswordHandler() {
	err := api.Queue.RegisterHandler(taskTypeResetPassword, asynq.HandlerFunc(api.handleResetPasswordTask))
	if err != nil {
		log.Errorf("register reset password handler err: %v", err)
	}
}

type ResetPasswordTask struct {
	TaskID         string    `json:"task_id"`
	RetryTimes     int       `json:"retry_times"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	RememberMe     bool      `json:"remember_me"`
	VerificationID string    `json:"verification_id"`
	EmailSendTime  time.Time `json:"email_send_time"`
}

func (api *API) handleResetPasswordTask(ctx context.Context, t *asynq.Task) error {
	task := &ResetPasswordTask{}
	if err := json.Unmarshal(t.Payload(), task); err != nil {
		return err
	}

	code, found, err := api.resetPasswordGetCode(ctx, task.Email, task.EmailSendTime)
	if err != nil || !found {
		return err
	}
	verificationToken, err := api.verifyResetCode(ctx, task.VerificationID, code)
	if err != nil {
		api.Redis.SetEx(ctx, task.TaskID, resetStatusError, time.Hour)
		return nil
	}

	err = api.resetPassword(ctx, task.Email, verificationToken, task.Password)
	if err != nil {
		api.Redis.SetEx(ctx, task.TaskID, resetStatusError, time.Hour)
		return nil
	}

	if task.RememberMe {
		ma := api.q.MasterAccount
		ma.WithContext(ctx).Where(ma.Email.Eq(task.Email)).Update(ma.Password, task.Password)
	}
	api.Redis.SetEx(ctx, task.TaskID, resetStatusDone, time.Hour)
	return nil
}
