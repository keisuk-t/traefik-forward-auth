package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/yudai/pp"
	"golang.org/x/oauth2"
)

// OIDC provider
type OIDC struct {
	IssuerURL    string `long:"issuer-url" env:"ISSUER_URL" description:"Issuer URL"`
	ClientID     string `long:"client-id" env:"CLIENT_ID" description:"Client ID"`
	ClientSecret string `long:"client-secret" env:"CLIENT_SECRET" description:"Client Secret" json:"-"`

	OAuthProvider

	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

// Name returns the name of the provider
func (o *OIDC) Name() string {
	return "oidc"
}

// Setup performs validation and setup
func (o *OIDC) Setup() error {
	// Check parms
	if o.IssuerURL == "" || o.ClientID == "" || o.ClientSecret == "" {
		return errors.New("providers.oidc.issuer-url, providers.oidc.client-id, providers.oidc.client-secret must be set")
	}

	var err error
	o.ctx = context.Background()

	// Try to initiate provider
	o.provider, err = oidc.NewProvider(o.ctx, o.IssuerURL)
	if err != nil {
		return err
	}

	// Create oauth2 config
	o.Config = &oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Endpoint:     o.provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Create OIDC verifier
	o.verifier = o.provider.Verifier(&oidc.Config{
		ClientID: o.ClientID,
	})

	return nil
}

// GetLoginURL provides the login url for the given redirect uri and state
func (o *OIDC) GetLoginURL(redirectURI, state string) string {
	return o.OAuthGetLoginURL(redirectURI, state)
}

// ExchangeCode exchanges the given redirect uri and code for a token
func (o *OIDC) ExchangeCode(redirectURI, code string) (string, error) {
	token, err := o.OAuthExchangeCode(redirectURI, code)
	if err != nil {
		return "", err
	}

	pp.Print("AccessToken: " + token.AccessToken)
	pp.Print("RefreshToken: " + token.RefreshToken)

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", errors.New("Missing id_token")
	}

	// Verify ID token
	idToken, err := o.verifier.Verify(o.ctx, rawIDToken)
	if err != nil {
		return "", err
	}

	pp.Print(idToken)

	var user User
	// Extract custom claims
	if err := idToken.Claims(&user); err != nil || len(user.Email) == 0 {
		// IDトークンにemailがなければ、userinfo APIを使って取得する
		userinfo, err := o.provider.UserInfo(o.ctx, o.Config.TokenSource(o.ctx, token))
		if err != nil {
			return "", err
		}

		pp.Print(userinfo)

		return rawIDToken + "|" + userinfo.Email, err
	}

	return rawIDToken + "|" + user.Email, nil
}

// GetUser uses the given token and returns a complete provider.User object
func (o *OIDC) GetUser(token string) (User, error) {
	elements := strings.Split(token, "|")
	if len(elements) > 1 {
		return User{Email: elements[1]}, nil
	}

	return User{}, errors.New("Missing email")
}
