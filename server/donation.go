package server

import (
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/utils"
	"net/http"
	"regexp"
	"strings"
)

var ppRedeemCodeReg = regexp.MustCompile(`^[-_0-9a-zA-Z]{16}$`)

func donationRedeemCode(c *gin.Context) {
	type Req struct {
		Nickname    string   `json:"nickname"`
		ChannelID   string   `json:"channel_id"`
		Drive       string   `json:"drive"`
		RedeemCodes []string `json:"redeem_codes"`
	}
	var req Req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hostName := strings.ToLower(req.Drive)
	if !utils.Contains([]string{"pikpak"}, hostName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drive"})
		return
	}

	if len(req.RedeemCodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no redeem codes"})
		return
	}

	nickname := req.Nickname
	if len(nickname) > 128 {
		nickname = req.Nickname[:128]
	}
	host := hosts.Get(hostName)
	if host == nil {
		log.Errorf("invalid host: %s", hostName)
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_host", i18n.WithDataMap("host", hostName)))
		return
	}
	channelID := strings.TrimSpace(req.ChannelID)

	userInfo, err := query.User.WithContext(c).Where(query.User.Channel.Eq(channelID)).First()
	if err != nil {
		log.Errorf("get user count err: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validate channel_id failed"})
		return
	}

	// filter invalid redeem codes
	redeemCodes := make([]string, 0)
	for _, redeemCode := range req.RedeemCodes {
		cleanCode := strings.TrimSpace(redeemCode)
		if ppRedeemCodeReg.MatchString(cleanCode) {
			redeemCodes = append(redeemCodes, cleanCode)
		}
	}

	if len(redeemCodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid redeem codes is empty"})
		return
	}

	err = host.DonateRedeemCode(c.Request.Context(), nickname, userInfo.ID, req.RedeemCodes)
	if err != nil {
		log.Errorf("donate redeem code err: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
