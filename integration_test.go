// +build integration

package ebay_test

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"github.com/jybp/ebay"
	"golang.org/x/oauth2"
	"os"
	"testing"
	"net/http"
	"strings"
	"fmt"
	"bytes"
	"time"
	"io/ioutil"
	"encoding/json"
	"net/url"
)

var client *ebay.Client

func init() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		panic("No CLIENT_ID or CLIENT_SECRET. Tests won't run.")
	}
	c := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.ReuseTokenSource(nil, TokenSource{
				Endpoint: "https://api.sandbox.ebay.com/identity/v1/oauth2/token",
				ID:       clientID,
				Secret:   clientSecret,
				Scopes:  []string{"https://api.ebay.com/oauth/api_scope"},
				Client:   http.DefaultClient,
			}),
			Base: http.DefaultTransport,
		},
	}
	client = ebay.NewSandboxClient(c)
}

type TokenSource struct {
	Endpoint string
	ID       string
	Secret   string
	Scopes	 []string
	Client   *http.Client
}

func (ts TokenSource) Token() (*oauth2.Token, error) {
	scopes := strings.Join(ts.Scopes, " ")
	req, err := http.NewRequest(http.MethodPost,
		ts.Endpoint,
		strings.NewReader(fmt.Sprintf("grant_type=client_credentials&scope=%s", url.PathEscape(scopes))))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(ts.ID, ts.Secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ts.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if c := resp.StatusCode; c < 200 || c >= 300 {
		return nil, fmt.Errorf("%s\nStatus:\n%d", req.URL, resp.StatusCode)
	}
	token := struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	if err = json.NewDecoder(bytes.NewReader(body)).Decode(&token); err != nil {
		return nil, err
	}
	t := oauth2.Token{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
	}
	if secs := token.ExpiresIn; secs > 0 {
		t.Expiry = time.Now().Add(time.Duration(secs) * time.Second)
	}
	print(t.TokenType)
	print("\n")
	print(t.AccessToken)
	print("\n")
	return &t, nil
}

func TestAuthorization(t *testing.T) {
	// TODO user token is reauired
	req, err := client.NewRequest("GET", "buy/browse/v1/item_summary/search?q=drone&limit=3")
	t.Log(req, err)
	into := map[string]string{}
	err = client.Do(context.Background(), req, &into)
	t.Log(into, err)
}
