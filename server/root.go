// Copyright 2023 The KeepShare Authors. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"github.com/KeepShareOrg/keepshare/server/constant"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/KeepShareOrg/keepshare/config"
	"github.com/KeepShareOrg/keepshare/hosts"
	"github.com/KeepShareOrg/keepshare/locale"
	"github.com/KeepShareOrg/keepshare/pkg/gormutil"
	"github.com/KeepShareOrg/keepshare/pkg/i18n"
	"github.com/KeepShareOrg/keepshare/pkg/log"
	q "github.com/KeepShareOrg/keepshare/pkg/queue"
	mdw "github.com/KeepShareOrg/keepshare/server/middleware"
	"github.com/KeepShareOrg/keepshare/server/query"
	"github.com/KeepShareOrg/keepshare/static"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

// Map alias.
type Map = map[string]any

// Start server.
func Start() error {
	// load config
	if err := config.Load(); err != nil {
		return fmt.Errorf("load config err: %w", err)
	}
	log.WithField("configs", viper.AllSettings()).Debug("viper all settings")

	// reset mysql logger
	config.MySQL().Logger = gormutil.GormLogger(config.LogLevel())
	// init default query
	query.SetDefault(config.MySQL())
	initQueryTypes()

	// init queue with redis, select another db = db + 1.
	redisOpt := *(config.Redis().Options())
	redisOpt.DB = (redisOpt.DB + 1) % 16
	queueIns := q.New(redisOpt, map[string]int{
		constant.AsyncQueueInviteSubAccount: 6,
		constant.AsyncQueueResetPassword:    6,
		constant.AsyncQueueSyncWorkerInfo:   3,
		constant.AsyncQueueRefreshToken:     3,
		constant.AsyncQueueStatisticTask:    1,
	})
	queue = queueIns.Client()
	queue.RegisterHandler(statisticTask, asynq.HandlerFunc(handleGetStatistics))

	// load locales
	if err := i18n.Load(locale.FS); err != nil {
		return fmt.Errorf("load locales err: %w", err)
	}
	log.Info("loaded languages:", i18n.Languages())

	hosts.Start(&hosts.Dependencies{
		Mysql:  config.MySQL(),
		Redis:  config.Redis(),
		Mailer: config.Mailer(),
		Queue:  queue,
	})

	queueIns.Run() // Run after hosts start.

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		gin.Recovery(),
		mdw.CORS(),
		mdw.AccessLogger(regexp.MustCompile("^/(api|session)/"), autoSharingPath),
	)

	// health check
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	sessionRouter(router)
	apiRouter(router)
	consoleRouter(router)

	router.NoRoute(autoRouter)

	srv := &http.Server{
		Addr:    config.ListenHTTP(),
		Handler: router,
	}

	go getAsyncBackgroundTaskInstance().run()
	return serveGraceful(srv)
}

func sessionRouter(router *gin.Engine) {
	g := router.Group("/session")
	g.Use(mdw.ContextWithAcceptLanguage)

	g.POST("/sign_up", signUp)
	g.POST("/sign_in", signIn)
	g.POST("/sign_out", signOut)
	g.POST("/token", refreshToken)
	g.GET("/me", mdw.Auth, getUserInfo)
}

func apiRouter(router *gin.Engine) {
	g := router.Group("/api")
	g.Use(mdw.ContextWithAcceptLanguage)

	g.GET("/verification", verifyAccount)
	g.POST("/verification", mdw.Auth, sendVerificationLink)
	g.POST("/change_email", mdw.Auth, changeAccountEmail)
	g.POST("/change_password", mdw.Auth, changeAccountPassword)
	g.POST("/send_verification_code", sendVerificationCode)
	g.POST("/reset_password", resetPassword)
	g.POST("/send_verification_link", mdw.Auth, sendVerificationLink)

	g.GET("/shared_link", querySharedLinkInfo) // front-end query shared link status, authentication is not required
	g.GET("/shared_links", mdw.Auth, listSharedLinks)
	g.POST("/shared_links", mdw.Auth, createSharedLinks)
	g.POST("/query_shared_links", mdw.Auth, batchQuerySharedLinksInfo)
	g.DELETE("/shared_links", mdw.Auth, deleteSharedLinks)

	g.POST("/storage/statistics", mdw.Auth, storageStatistics)
	g.POST("/storage/release", mdw.Auth, storageRelease)

	g.GET("/blacklist", mdw.Auth, getBlackList)
	g.POST("/blacklist", mdw.Auth, addToBlackList)
	g.DELETE("/blacklist", mdw.Auth, removeFromBlackList)

	g.GET("/host/info", mdw.Auth, getHostInfo)

	g.PATCH("/host/password", mdw.Auth, changeHostPassword)
	g.GET("/host/password/task", mdw.Auth, getChangePasswordTaskInfo)
}

func consoleRouter(router *gin.Engine) {
	if config.ConsoleProxyURL() != "" {
		router.GET("/console/*path", consoleProxy())
		return
	}

	sub, err := fs.Sub(static.FS, "dist")
	// frontend pages.
	if err != nil {
		panic(err)
		return
	}
	h := http.FileServer(http.FS(sub))

	indexPage, err := static.FS.ReadFile("dist/console/index.html")
	if err != nil {
		panic(err)
	}

	router.GET("/console/*path", func(c *gin.Context) {
		path := c.Request.URL.Path
		c.Set(mdw.SkipAccessLog, false)

		switch {
		case strings.HasPrefix(path, "/console/assets"):
			fallthrough
		case path == "/console/logo.ico":
			fallthrough
		case path == "/console/index.html":
			h.ServeHTTP(&staticWriter{c.Writer}, c.Request)
		default:
			if err != nil {
				return
			}
			c.Data(http.StatusOK, "text/html", indexPage)
		}
	})
}

var mockViteContentType = regexp.MustCompile(`(?i)\.(json|png)`)

func consoleProxy() func(c *gin.Context) {
	proxyURL := config.ConsoleProxyURL()
	if strings.HasSuffix(proxyURL, "/") {
		proxyURL = proxyURL[:len(proxyURL)-1]
	}
	handler := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			link, _ := url.Parse(proxyURL + request.URL.Path + "?" + request.URL.RawQuery)
			request.URL = link
		},
		ModifyResponse: func(response *http.Response) error {
			if mockViteContentType.MatchString(response.Request.URL.Path) {
				response.Header.Set("Content-Type", "application/javascript")
			}
			return nil
		},
	}
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

var autoSharingPath = regexp.MustCompile(`(?i)^/[a-z0-9]{8}/(magnet|http|https|ftp|ed2k):`)

func autoRouter(c *gin.Context) {
	switch {
	case c.Request.Method == http.MethodGet && autoSharingPath.MatchString(c.Request.URL.Path):
		// get or create a shared link from original link
		mdw.ContextWithAcceptLanguage(c)
		autoSharingLink(c)
	default:
		c.Redirect(http.StatusFound, "/")
	}
}

func serveGraceful(srv *http.Server) error {
	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Errorf("listen err: %v", err)
			errChan <- err
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err

	case <-quit:
		log.Warn("shutting down server")

		// The context is used to inform the server it has 5 seconds to finish
		// the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Warn("shutdown server err:", err)
		}

		log.Warn("server stopped")
		return nil
	}
}

type staticWriter struct {
	http.ResponseWriter
}

func (w *staticWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusOK {
		if w.Header().Get("Cache-Control") == "" {
			// static files cache one month
			w.Header().Set("Cache-Control", "public, max-age=2592000")
		}
	}
	w.ResponseWriter.WriteHeader(statusCode)
}
