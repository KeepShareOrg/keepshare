// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"strconv"
	"time"
)

func sendVerificationLink(c *gin.Context) {
	ctx := c.Request.Context()
	userId := c.GetString(constant.UserID)

	user, err := query.User.WithContext(ctx).Where(query.User.ID.Eq(userId)).Take()
	if err != nil {
		c.JSON(http.StatusBadGateway, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	salt := viper.Get("verify_email_link_salt")
	expires := viper.Get("verify_email_link_expires")
	expiresDuration, err := time.ParseDuration(expires.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}
	expiresTime := time.Now().Add(expiresDuration).Unix()

	verifyString := fmt.Sprintf("%v-%v-%v-%v", user.Email, user.ID, expiresTime, salt)
	hash := CalcSha265Hash(verifyString, salt.(string))
	verifyLink := fmt.Sprintf("https://%v/api/verification?token=%v&email=%v&expires=%v", config.RootDomain(), hash, user.Email, expiresTime)
	log.WithContext(ctx).Debugf("verify link: %s", verifyLink)

	emailClient, err := GetEmailClient()
	if err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	emailHTMLContent := fmt.Sprintf(viper.GetString("confirm_email_html_template"), verifyLink)
	emailTextContent := fmt.Sprintf(viper.GetString("confirm_email_text_template"), verifyLink)
	if err := emailClient.NewMessage("KeepShare - Verify your email").
		AddHtmlContent(emailHTMLContent).
		AddTextContent(emailTextContent).
		Send([]string{user.Email}); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func verifyAccount(c *gin.Context) {
	ctx := c.Request.Context()

	resultPageAddr := fmt.Sprintf("https://%v/console/email-verification", config.RootDomain())
	successAddr := fmt.Sprintf("%v?success=1", resultPageAddr)
	failedAddr := fmt.Sprintf("%v?success=0", resultPageAddr)
	expiresAddr := fmt.Sprintf("%v?expired=1", resultPageAddr)

	email := c.Query("email")
	token := c.Query("token")
	expiresTime := c.Query("expires")
	if email == "" || token == "" || expiresTime == "" {
		c.Redirect(http.StatusMovedPermanently, failedAddr)
		return
	}

	expiresUnix, err := strconv.ParseInt(expiresTime, 10, 64)
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, failedAddr)
		return
	}

	if time.Now().Unix() > expiresUnix {
		c.Redirect(http.StatusMovedPermanently, expiresAddr)
		return
	}

	user, err := query.User.WithContext(ctx).Where(query.User.Email.Eq(email)).Take()
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, failedAddr)
		return
	}

	salt := viper.Get("verify_email_link_salt")
	verifyString := fmt.Sprintf("%v-%v-%v-%v", user.Email, user.ID, expiresTime, salt)
	hash := CalcSha265Hash(verifyString, salt.(string))

	if token != hash {
		c.Redirect(http.StatusMovedPermanently, failedAddr)
		return
	}

	if _, err = query.User.
		WithContext(ctx).
		Where(query.User.ID.Eq(user.ID)).
		Update(query.User.EmailVerified, constant.EmailVerificationDone); err != nil {
		c.Redirect(http.StatusMovedPermanently, failedAddr)
		return
	}

	c.Redirect(http.StatusFound, successAddr)
}
