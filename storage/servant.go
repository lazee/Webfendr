package storage

import (
	"bufio"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"webfendr/authenticator"
	"webfendr/config"
)

func FileHandler(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		siteHost := strings.Split(ctx.Request.Host, ":")[0]
		uri := extractUri(ctx)

		log.Debug("Servant: siteHost: ", siteHost, ", uri: ", uri)
		// Support front page for Webster itself, when direct request to webfendr domain
		if ctx.Request.Host == cfg.WebFendrHost && uri == "/index.html" {
			ctx.HTML(http.StatusOK, "webfendr.html", gin.H{
				"domains": authenticator.FindDomains(cfg),
			})
			return
		}

		filePath := createFilePath(cfg, siteHost, uri)
		file, err := os.Open(filePath)
		if err != nil {
			log.Debug("Not found: ", uri)
			ctx.HTML(http.StatusNotFound, "404.html", nil)
			return
		}
		fileReader := bufio.NewReader(file)
		mimeType, err := mimetype.DetectFile(filePath)
		ctx.DataFromReader(
			http.StatusOK,
			-1,
			mimeType.String(),
			fileReader, nil)
		defer file.Close()
	}

}

func createFilePath(cfg *config.Config, webFendrHost string, uri string) string {
	return filepath.Join(cfg.SiteDir, webFendrHost, uri)
}

func extractUri(ctx *gin.Context) string {
	uri := strings.Split(ctx.Request.RequestURI, "?")[0]
	if strings.HasSuffix(uri, "/") {
		uri += "index.html"
	}
	return uri
}
