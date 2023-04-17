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
		// Support front page for Webfendr itself, when direct request to webfendr domain
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
		mimeType := detectMime(filePath)
		log.Info("mimetype", mimeType)
		ctx.DataFromReader(
			http.StatusOK,
			-1,
			mimeType,
			fileReader, nil)
		defer file.Close()
	}

}

// Hack since mimetype package is stupid
func detectMime(filePath string) string {
	mimeType, err := mimetype.DetectFile(filePath)
	if err != nil {
		log.Error("Could not detect mime type for ", filePath)
		return "text/plain"
	}
	mimeStr := mimeType.String()
	if strings.Contains(mimeStr, "text/plain") {
		if strings.HasSuffix(filePath, ".css") {
			return "text/css; charset=utf-8"
		} else if strings.HasSuffix(filePath, ".js") {
			return "text/javascript; charset=utf-8"
		} else if strings.HasSuffix(filePath, ".html") {
			return "text/html; charset=utf-8"
		} else if strings.HasSuffix(filePath, ".woff") {
			return "font/woff"
		} else if strings.HasSuffix(filePath, ".woff2") {
			return "font/woff2"
		} else if strings.HasSuffix(filePath, ".png") {
			return "image/png"
		} else if strings.HasSuffix(filePath, ".jpg") {
			return "image/jpg"
		} else if strings.HasSuffix(filePath, ".jpeg") {
			return "image/jpeg"
		} else if strings.HasSuffix(filePath, ".ico") {
			return "image/vnd.microsoft.icon"
		}
	}
	return mimeStr
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
