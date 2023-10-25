package server

import (
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/server/constant"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/gin-gonic/gin"
	"net/http"
)

func changeAccountEmail(c *gin.Context) {
	type Req struct {
		NewEmail     string `json:"new_email"`
		PasswordHash string `json:"password_hash"`
	}

	resp := gin.H{"success": true}

	var req Req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	if req.NewEmail == c.GetString(constant.Email) {
		c.JSON(http.StatusOK, resp)
		return
	}

	ctx := c.Request.Context()
	userId := c.GetString(constant.UserID)
	user, err := query.User.WithContext(ctx).Where(query.User.ID.Eq(userId)).Take()
	if err != nil {
		c.JSON(http.StatusBadGateway, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	if user.PasswordHash != req.PasswordHash {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid password")))
		return
	}

	if _, err := query.User.
		WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		Update(query.User.Email, req.NewEmail); err != nil {
		c.JSON(http.StatusBadGateway, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func changeAccountPassword(c *gin.Context) {
	type Req struct {
		PasswordHash    string `json:"password_hash"`
		NewPasswordHash string `json:"new_password_hash"`
	}

	resp := gin.H{"success": true}

	var req Req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", err.Error())))
		return
	}

	ctx := c.Request.Context()
	userId := c.GetString(constant.UserID)
	user, err := query.User.WithContext(ctx).Where(query.User.ID.Eq(userId)).Take()
	if err != nil {
		c.JSON(http.StatusBadGateway, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	if user.PasswordHash != req.PasswordHash {
		c.JSON(http.StatusBadRequest, mdw.ErrResp(c, "invalid_params", i18n.WithDataMap("error", "invalid password")))
		return
	}

	if _, err := query.User.
		WithContext(ctx).
		Where(query.User.ID.Eq(userId)).
		Update(query.User.PasswordHash, req.NewPasswordHash); err != nil {
		c.JSON(http.StatusBadGateway, mdw.ErrResp(c, "internal", i18n.WithDataMap("error", err.Error())))
		return
	}

	c.JSON(http.StatusOK, resp)
}
