package auth

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/flimzy/kivik"
	"github.com/go-kivik/couchdb/chttp"
)

// NewOAuth2 returns a new kivik chttp authenticator based on the provided
// OAuth2 provider and token.
func NewOAuth2(provider, token string) *OAuth2Authenticator {
	return &OAuth2Authenticator{
		Provider: provider,
		Token:    token,
	}
}

// OAuth2Authenticator allows chttp authentication with Flashback's OAuth2
// middleware proxy.
type OAuth2Authenticator struct {
	Provider string `json:"provider"`
	Token    string `json:"access_token"`
}

var _ chttp.Authenticator = &OAuth2Authenticator{}

// Authenticate attempts to authenticate with an OAuth2 token.
func (a *OAuth2Authenticator) Authenticate(ctx context.Context, client *chttp.Client) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(a); err != nil {
		return err
	}
	var result struct {
		Name string `json:"name"`
	}
	if _, err := client.DoJSON(ctx, kivik.MethodPost, "/_session", &chttp.Options{Body: buf}, &result); err != nil {
		return err
	}
	return chttp.ValidateAuth(ctx, result.Name, client)
}

// Logout logs out.
func (a *OAuth2Authenticator) Logout(ctx context.Context, client *chttp.Client) error {
	return nil
}
