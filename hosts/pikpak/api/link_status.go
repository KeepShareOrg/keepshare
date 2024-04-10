package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/pkg/log"
)

// QueryLinkStatus query link status
func (api *API) QueryLinkStatus(ctx context.Context, link string) (statue string, progress int) {
	var r struct {
		Status   string `json:"status"`
		Progress int    `json:"progress"`
	}

	resp, err := http.Get(apiURL(fmt.Sprintf("/drive/v1/resource/status?url=%s", url.QueryEscape(link))))
	if err != nil {
		log.Debugf("query link status err: %v", err)
		return comm.LinkStatusUnknown, 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("read link status response err: %v", err)
		return comm.LinkStatusUnknown, 0
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Debugf("unmarshal link status response err: %v", err)
		return comm.LinkStatusUnknown, 0
	}

	log.WithContext(ctx).Debugf("query link status resp body: %#v", r)

	return r.Status, r.Progress
}
