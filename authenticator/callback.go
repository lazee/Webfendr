package authenticator

import (
	"net/http"
	"webfendr/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CallbackHandler(auths *Authenticators, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		if ctx.Query("state") != session.Get("state") {
			ctx.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		redirectUri, ok := session.Get("redirectUri").(string)
		if !ok {
			redirectUri = "/"
		}

		auth, err := auths.Get(ctx, cfg)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		token, err := auth.Exchange(ctx.Request.Context(), ctx.Query("code"))
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
			return
		}

		idToken, err := auth.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		session.Set("access_token", token.AccessToken)
		session.Set("profile", profile)
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Redirect(http.StatusTemporaryRedirect, redirectUri)
	}
}
