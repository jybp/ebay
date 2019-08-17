package integration

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jybp/ebay"
	"github.com/jybp/ebay/tokensource"
	"golang.org/x/oauth2"
	oclientcredentials "golang.org/x/oauth2/clientcredentials"
)

var (
	integration  bool
	clientID     string
	clientSecret string
)

func init() {
	flag.BoolVar(&integration, "integration", false, "run integration tests")
	flag.Parse()
	if !integration {
		return
	}
	clientID = os.Getenv("SANDBOX_CLIENT_ID")
	clientSecret = os.Getenv("SANDBOX_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		panic("No SANDBOX_CLIENT_ID or SANDBOX_CLIENT_SECRET. Tests won't run.")
	}
}

func TestAuction(t *testing.T) {
	if !integration {
		t.SkipNow()
	}

	// Manually create an auction in the sandbox and copy/paste the url.
	// Auctions can't be created using the rest api (yet?).
	const auctionURL = "https://www.sandbox.ebay.com/itm/110440008951"

	ctx := context.Background()

	conf := oclientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://api.sandbox.ebay.com/identity/v1/oauth2/token",
		Scopes:       []string{"https://api.ebay.com/oauth/api_scope"},
	}

	client := ebay.NewSandboxClient(oauth2.NewClient(ctx, tokensource.New(conf.TokenSource(ctx))))

	lit, err := client.Buy.Browse.GetItemByLegacyID(ctx, auctionURL[strings.LastIndex(auctionURL, "/")+1:])
	if err != nil {
		t.Fatalf("%+v", err)
	}
	it, err := client.Buy.Browse.GetItem(ctx, lit.ItemID)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if testing.Verbose() {
		t.Logf("item: %+v\n", it)
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
	t.Logf("item %s UniqueBidderCount:%d minimumBidPrice: %+v currentPriceToBid: %+v\n", it.ItemID, it.UniqueBidderCount, it.MinimumPriceToBid, it.CurrentBidPrice)

	// Setup oauth server.

	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		t.Fatalf("%+v", err)
	}
	state := url.QueryEscape(string(b))
	authCodeC := make(chan string)
	http.HandleFunc("/accept", func(rw http.ResponseWriter, r *http.Request) {
		actualState, err := url.QueryUnescape(r.URL.Query().Get("state"))
		if err != nil {
			http.Error(rw, fmt.Sprintf("invalid state: %+v", err), http.StatusBadRequest)
			return
		}
		if string(actualState) != state {
			http.Error(rw, fmt.Sprintf("invalid state:\nexpected:%s\nactual:%s", state, string(actualState)), http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		authCodeC <- code
		t.Logf("The authorization code is %s.\n", code)
		t.Logf("The authorization code will expire in %s seconds.\n", r.URL.Query().Get("expires_in"))
		rw.Write([]byte("Accept. You can safely close this tab."))
	})
	http.HandleFunc("/policy", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("eBay Sniper Policy"))
	})
	http.HandleFunc("/decline", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Decline. You can safely close this tab."))
	})
	go func() {
		t.Fatal(http.ListenAndServe(":52125", nil))
	}()

	oauthConf := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.sandbox.ebay.com/oauth2/authorize",
			TokenURL: "https://api.sandbox.ebay.com/identity/v1/oauth2/token",
		},
		RedirectURL: "Jean-Baptiste_P-JeanBapt-testgo-cowrprk",
		Scopes:      []string{"https://api.ebay.com/oauth/api_scope/buy.offer.auction"},
	}

	url := oauthConf.AuthCodeURL(state)
	fmt.Printf("Visit the URL: %v\n", url)

	authCode := <-authCodeC

	tok, err := oauthConf.Exchange(ctx, authCode)
	if err != nil {
		t.Fatal(err)
	}

	client = ebay.NewSandboxClient(oauth2.NewClient(ctx, tokensource.New(oauthConf.TokenSource(ctx, tok))))

	_, err = client.Buy.Offer.GetBidding(ctx, it.ItemID, ebay.BuyMarketplaceUSA)
	if !ebay.IsError(err, ebay.ErrGetBiddingNoBiddingActivity) {
		t.Logf("Expected ErrNoBiddingActivity, got %+v.", err)
	}

	// err := client.Buy.Offer.PlaceProxyBid(ctx)

}
