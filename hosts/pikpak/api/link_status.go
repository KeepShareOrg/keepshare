package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KeepShareOrg/keepshare/hosts/pikpak/comm"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"io"
	"net/http"
	"net/url"
)

// QueryLinkStatus query link status
func (api *API) QueryLinkStatus(ctx context.Context, link string) string {
	var r struct {
		Status string `json:"status"`
	}

	resp, err := http.Get(fmt.Sprintf("https://api-drive.mypikpak.com/drive/v1/resource/status?url=%s", url.PathEscape(link)))
	if err != nil {
		log.Debugf("query link status err: %v", err)
		return comm.LinkStatusUnknown
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("read link status response err: %v", err)
		return comm.LinkStatusUnknown
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Debugf("unmarshal link status response err: %v", err)
		return comm.LinkStatusUnknown
	}

	log.WithContext(ctx).Debugf("query link status resp body: %#v", r)

	return r.Status
}
