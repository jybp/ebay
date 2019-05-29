// +build integration
package integration

import (
	"context"
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jybp/ebay"
	"github.com/jybp/ebay/clientcredentials"
)

var client *ebay.Client

func init() {
	clientID := os.Getenv("SANDBOX_CLIENT_ID")
	clientSecret := os.Getenv("SANDBOX_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		panic("No SANDBOX_CLIENT_ID or SANDBOX_CLIENT_SECRET. Tests won't run.")
	}
	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://api.sandbox.ebay.com/identity/v1/oauth2/token",
		Scopes:       []string{"https://api.ebay.com/oauth/api_scope"},
	}
	client = ebay.NewSandboxClient(conf.Client(context.Background()))
}

func TestAuthorization(t *testing.T) {
	// https://developer.ebay.com/my/api_test_tool?index=0&api=browse&call=item_summary_search__GET&variation=json
	req, err := client.NewRequest("GET", "buy/browse/v1/item_summary/search?q=test")
	if err != nil {
		t.Fatal(err)
	}
	into := map[string]interface{}{}
	err = client.Do(context.Background(), req, &into)
	if err != nil {
		t.Fatal(err)
	}
	if testing.Verbose() {
		t.Log(into)
	}
}
