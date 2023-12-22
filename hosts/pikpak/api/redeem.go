package api

import (
	"context"
	"fmt"
	"github.com/KeepShareOrg/keepshare/pkg/log"
)

// Redeem redeem a redeem code and change the user's status
func (api *API) Redeem(ctx context.Context, userID, redeemCode string) error {
	token, err := api.getToken(ctx, userID, false)
	if err != nil {
		return err
	}

	var e RespErr
	var r struct {
		Code             string `json:"code"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	body, err := resCli.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetError(&e).
		SetResult(&r).
		SetBody(JSON{
			"activation_code": redeemCode,
		}).
		Post(apiURL("/vip/v1/order/activation-code"))

	if err != nil {
		return fmt.Errorf("redeem err: %w", err)
	}

	log.WithContext(ctx).Debugf("redeem resp body: %s", body.Body())

	if r.Error != "" {
		return fmt.Errorf("redeem err: %s", r.Error)
	}

	return nil
}
