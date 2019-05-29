// Package clientcredentials implements the eBay client credentials grant flow
// to generate "Application access" tokens.
// eBay doc: https://developer.ebay.com/api-docs/static/oauth-client-credentials-grant.html
//
// The only difference from the original golang.org/x/oauth2/clientcredentials is the token type
// being forced to "Bearer". The eBay api /identity/v1/oauth2/token endpoint returns
// "Application Access Token" as token type which is then reused:
// https://github.com/golang/oauth2/blob/aaccbc9213b0974828f81aaac109d194880e3014/token.go#L68-L70
package clientcredentials

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type Config struct {
	// ClientID is the application's ID.
	ClientID string

	// ClientSecret is the application's secret.
	ClientSecret string

	// TokenURL is the resource server's token endpoint
	// URL. This is a constant specific to each server.
	TokenURL string

	// Scope specifies optional requested permissions.
	Scopes []string

	// HTTPClient used to make requests.
	HTTPClient *http.Client
}

// Token uses client credentials to retrieve a token.
func (c *Config) Token(ctx context.Context) (*oauth2.Token, error) {
	return c.TokenSource(ctx).Token()
}

// Client returns an HTTP client using the provided token.
// The token will auto-refresh as necessary.
// The returned client and its Transport should not be modified.
func (c *Config) Client(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx))
}

// TokenSource returns a TokenSource that returns t until t expires,
// automatically refreshing it as necessary using the provided context and the
// client ID and client secret.
//
// Most users will use Config.Client instead.
func (c *Config) TokenSource(ctx context.Context) oauth2.TokenSource {
	source := &tokenSource{
		ctx:  ctx,
		conf: c,
	}
	return oauth2.ReuseTokenSource(nil, source)
}

type tokenSource struct {
	ctx  context.Context
	conf *Config
}

// Token refreshes the token by using a new client credentials request.
// tokens received this way do not include a refresh token
func (c *tokenSource) Token() (*oauth2.Token, error) {
	v := url.Values{
		"grant_type": {"client_credentials"},
	}
	if len(c.conf.Scopes) > 0 {
		v.Set("scope", strings.Join(c.conf.Scopes, " "))
	}
	req, err := http.NewRequest("POST", c.conf.TokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(url.QueryEscape(c.conf.ClientID), url.QueryEscape(c.conf.ClientSecret))
	client := c.conf.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch token: %v", err)
	}
	if code := r.StatusCode; code < 200 || code > 299 {
		return nil, fmt.Errorf("%s (%d)", req.URL, r.StatusCode)
	}
	token := struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	if err = json.NewDecoder(bytes.NewReader(body)).Decode(&token); err != nil {
		return nil, err
	}
	var expiry time.Time
	if secs := token.ExpiresIn; secs > 0 {
		expiry = time.Now().Add(time.Duration(secs) * time.Second)
	}
	t := oauth2.Token{
		AccessToken: token.AccessToken,
		Expiry:      expiry,
		TokenType:   "Bearer",
	}
	return &t, nil
}
