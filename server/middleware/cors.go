// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS creates a new CORS Gin middleware.
func CORS() gin.HandlerFunc {
	conf := cors.Config{
		AllowAllOrigins: true,
		AllowMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"accept",
			"authorization",
			"cache-control",
			"content-type",
			"csrf-token",
			"keep-alive",
			"origin",
			"user-agent",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(conf)
}
