// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ContextWithAcceptLanguage store accept language into c.Request.Context.
func ContextWithAcceptLanguage(c *gin.Context) {
	a := c.Request.Header.Get("Accept-Language")
	if a != "" {
		newCtx := i18n.ContextWithAcceptLanguage(c.Request.Context(), a)
		c.Request = c.Request.WithContext(newCtx)
	}

	//c.Next()
}
