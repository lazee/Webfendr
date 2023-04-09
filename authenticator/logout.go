package authenticator

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"webfendr/config"

	"github.com/gin-gonic/gin"
)

func LogoutHandler(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logoutUrl, err := url.Parse("https://" + cfg.Auth0Domain + "/v2/logout")
		if err != nil {
			log.Debug("Could not create logout url ", err.Error())
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		log.Debug("logoutUrl ", logoutUrl)

		returnTo, err := url.Parse(cfg.HttpProtocol() + "://" + ctx.Request.Host + "/webfendr")
		if err != nil {
			log.Debug("Could not returnTo url ", err.Error())
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		parameters := url.Values{}
		parameters.Add("returnTo", returnTo.String())
		parameters.Add("client_id", cfg.Auth0ClientId)
		logoutUrl.RawQuery = parameters.Encode()

		log.Debug("Redirect to ", logoutUrl.String())
		ctx.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
	}
}
