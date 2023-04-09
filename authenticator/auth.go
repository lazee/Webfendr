package authenticator

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"webfendr/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Authenticators struct {
	auths map[string]*Authenticator
}

func (a *Authenticators) Get(ctx *gin.Context, cfg *config.Config) (*Authenticator, error) {
	authenticator, ok := a.auths[ctx.Request.Host]
	if ok {
		log.Trace("Found Authenticator for ", ctx.Request.Host)
		return authenticator, nil
	}
	log.Trace("Creating new Authenticator for ", ctx.Request.Host)
	authenticator, err := NewAuthenticator(ctx, cfg)
	if err != nil {
		return nil, err
	}
	a.auths[ctx.Request.Host] = authenticator
	return authenticator, nil
}

func Init() *Authenticators {
	return &Authenticators{
		auths: make(map[string]*Authenticator),
	}
}

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

func CreateCallbackUrl(ctx *gin.Context, cfg *config.Config) string {
	scheme := "http"
	if cfg.Tls {
		scheme = "https"
	}
	return scheme + "://" + ctx.Request.Host + "/webfendr/callback"
}

func NewAuthenticator(ctx *gin.Context, cfg *config.Config) (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+cfg.Auth0Domain+"/",
	)
	if err != nil {
		return nil, err
	}

	log.Error("Creating authenticator with ", CreateCallbackUrl(ctx, cfg))
	conf := oauth2.Config{
		ClientID:     cfg.Auth0ClientId,
		ClientSecret: cfg.Auth0ClientSecret,
		RedirectURL:  CreateCallbackUrl(ctx, cfg),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
	}, nil
}

func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}
