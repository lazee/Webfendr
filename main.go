package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"webfendr/authenticator"
	"webfendr/config"
	"webfendr/storage"
)
import "github.com/gin-gonic/gin"

func BuildUrl(ctx *gin.Context, cfg *config.Config) string {
	return cfg.HttpProtocol() + "://" + ctx.Request.Host + ctx.Request.RequestURI
}

func IsAuthenticated(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if sessions.Default(ctx).Get("profile") == nil {
			session := sessions.Default(ctx)
			session.Set("redirectUri", BuildUrl(ctx, cfg))
			session.Save()
			log.Debug("Redirecting to /webfendr/login")
			ctx.Redirect(http.StatusSeeOther, "/webfendr/login")
		} else {
			ctx.Next()
		}
	}
}

func Router(auths *authenticator.Authenticators, cfg *config.Config, logger *log.Logger) *gin.Engine {
	router := gin.New()
	//router.Use(ginLogrus.Logger(logger), gin.Recovery(), gzip.Gzip(gzip.DefaultCompression))
	router.Use(gin.Recovery(), gzip.Gzip(gzip.DefaultCompression))
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.Static("/webfendr/public", filepath.Join(cfg.WebFolder, "web/static"))
	router.LoadHTMLGlob(filepath.Join(cfg.WebFolder, "web/template/**.html"))

	router.GET("/webfendr", authenticator.IndexHandler(cfg))
	router.GET("/webfendr/login", authenticator.LoginHandler(auths, cfg))
	router.GET("/webfendr/callback", authenticator.CallbackHandler(auths, cfg))
	router.GET("/webfendr/logout", authenticator.LogoutHandler(cfg))
	router.NoRoute(IsAuthenticated(cfg), storage.FileHandler(cfg))
	return router
}

func main() {
	// Load config
	cfg := config.PrepareConfig()

	// Init logging
	logrus := log.New()
	if cfg.WebFendrMode == gin.ReleaseMode {
		logrus.SetFormatter(&log.JSONFormatter{})
	} else {
		logrus.SetFormatter(&log.TextFormatter{})
	}
	gin.SetMode(cfg.WebFendrMode)

	// Init the oauth authenticator
	auths := authenticator.Init()

	// Create a default context
	ctx, cancel := context.WithCancel(context.Background())

	// Start the Google Storage synchronization runner
	go storage.Syncer(ctx, cfg)

	// Load Gin routes
	rtr := Router(auths, cfg, logrus)

	// Create server address
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	// Start http server
	log.Info("Listening on ", addr)
	if err := http.ListenAndServe(addr, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}

	// Run cancellation on default context
	cancel()
}
