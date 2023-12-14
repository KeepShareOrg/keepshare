package api

import (
	"context"
	"fmt"

	"github.com/KeepShareOrg/keepshare/pkg/log"
)

// JoinReferral join https://mypikpak.com/referral/
func (api *API) JoinReferral(ctx context.Context, userID string) error {
	token, err := api.getToken(ctx, userID, true)
	if err != nil {
		return err
	}

	var e RespErr
	var r struct {
		ID string `json:"id"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetBody(JSON{}).
		Post(referralURL("/promoting/v1/join"))

	if err != nil {
		return fmt.Errorf("join referral err: %w", err)
	}

	log.WithContext(ctx).WithField("user_id", userID).Debugf("join referral response body: %s", body.Body())

	if err = e.Error(); err != nil {
		// TODO token expired
		return fmt.Errorf("join referral err: %w", err)
	}

	if r.ID == "" {
		return fmt.Errorf("join referral got unexpected body: %s", body.Body())
	}
	return nil
}

// InviteSubAccount invite sub-account by email.
func (api *API) InviteSubAccount(ctx context.Context, masterUserID string, workerEmail string) error {
	token, err := api.getToken(ctx, masterUserID, true)
	if err != nil {
		return err
	}

	var e RespErr
	var r struct{}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetBody(JSON{"email": workerEmail}).
		Post(referralURL("/promoting/v1/sub-account"))

	if err != nil {
		return fmt.Errorf("invite sub account err: %w", err)
	}

	log.WithContext(ctx).WithField("user_id", masterUserID).Debugf("invite sub account response body: %s", body.Body())

	if err = e.Error(); err != nil {
		// TODO token expired
		return fmt.Errorf("invite sub account err: %w", err)
	}

	return nil
}

// VerifyInviteSubAccountToken invite sub-account by email.
func (api *API) VerifyInviteSubAccountToken(ctx context.Context, token string) error {
	var e RespErr
	var r struct{}

	body, err := resCli.R().
		SetContext(ctx).
		SetQueryParam("token", token).
		SetError(&e).
		SetResult(&r).
		Get(referralURL("/promoting/v1/sub-account/verify"))

	if err != nil {
		return fmt.Errorf("verify invite sub account token err: %w", err)
	}

	log.WithContext(ctx).Debugf("verify invite sub account token response body: %s", body.Body())

	if err = e.Error(); err != nil {
		return fmt.Errorf("verify invite sub account token err: %w", err)
	}

	return nil
}

// GetCommissionsResponse is the response of the GetCommissions API.
type GetCommissionsResponse struct {
	Total     float64 `json:"total"`
	Pending   float64 `json:"pending"`
	Available float64 `json:"available"`
}

// GetCommissions get commissions from server.
func (api *API) GetCommissions(ctx context.Context, userID string) (*GetCommissionsResponse, error) {
	token, err := api.getToken(ctx, userID, true)
	if err != nil {
		return nil, err
	}

	var e RespErr
	var r GetCommissionsResponse

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		Get(referralURL("/promoting/v1/commissions/summary"))

	if err != nil {
		return nil, fmt.Errorf("get commissions err: %w", err)
	}

	log.WithContext(ctx).WithField("user_id", userID).Debugf("get commissions response body: %s", body.Body())

	if err = e.Error(); err != nil {
		// TODO token expired
		return nil, fmt.Errorf("get commissions err: %w", err)
	}

	return &r, nil
}
