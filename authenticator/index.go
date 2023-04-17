package authenticator

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"webfendr/config"
)

type IndexData struct {
	authenticated bool
	profile       map[string]interface{}
}

func FindDomains(cfg *config.Config) []string {
	var a []string
	f, err := os.Open(cfg.SiteDir)
	if err != nil {
		log.Error(err)
		return a
	}
	files, err := f.Readdir(0)
	if err != nil {
		log.Error(err)
		return a
	}

	for _, v := range files {
		if !v.IsDir() && strings.HasSuffix(v.Name(), ".zip") {
			a = append(a, cfg.HttpProtocol()+"://"+strings.TrimSuffix(v.Name(), ".zip"))
		}
	}
	return a
}

func IndexHandler(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		if ctx.Request.Host == cfg.WebFendrHost {
			ctx.Header("Cache-Control", "no-cache")
			ctx.HTML(http.StatusOK, filepath.Join(cfg.WebFolder, "webfendr.html"), gin.H{
				"domains": FindDomains(cfg),
			})
			return
		}

		session := sessions.Default(ctx)
		profile := session.Get("profile")

		ctx.Header("Cache-Control", "no-cache")
		ctx.HTML(http.StatusOK, filepath.Join(cfg.WebFolder, "index.html"), gin.H{
			"authenticated": sessions.Default(ctx).Get("profile") != nil,
			"profile":       profile,
		})
	}
}
