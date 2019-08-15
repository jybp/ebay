package integration

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jybp/ebay"
	"github.com/jybp/ebay/clientcredentials"
)

var (
	integration bool
	client      *ebay.Client
)

func init() {
	flag.BoolVar(&integration, "integration", false, "run integration tests")
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

func TestAuction(t *testing.T) {
	if !integration {
		t.SkipNow()
	}

	// Manually create an auction in the sandbox and copy/paste the url:
	const url = "https://www.sandbox.ebay.com/itm/110439278158"

	ctx := context.Background()
	lit, err := client.Buy.Browse.GetItemByLegacyID(ctx, url[strings.LastIndex(url, "/")+1:])
	if err != nil {
		t.Fatalf("%+v", err)
	}
	it, err := client.Buy.Browse.GetItem(ctx, lit.ItemID)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("item: %+v", it)
	}
	isAuction := false
	for _, opt := range it.BuyingOptions {
		if opt == ebay.BrowseBuyingOptionAuction {
			isAuction = true
		}
	}
	if !isAuction {
		t.Fatalf("item %s is not an auction. BuyingOptions are: %+v", it.ItemID, it.BuyingOptions)
	}
	if time.Now().UTC().After(it.ItemEndDate) {
		t.Fatalf("item %s end date has been reached. ItemEndDate is: %s", it.ItemID, it.ItemEndDate.String())
	}
	t.Logf("item %s UniqueBidderCount:%d minimumBidPrice: %+v currentPriceToBid: %+v", it.ItemID, it.UniqueBidderCount, it.MinimumPriceToBid, it.CurrentBidPrice)
}
