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
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/jybp/ebay"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	integration  bool
	clientID     string
	clientSecret string
	auctionURL   string
)

func init() {
	flag.BoolVar(&integration, "integration", false, "run integration tests")
	flag.Parse()
	if !integration {
		return
	}
	clientID = os.Getenv("SANDBOX_CLIENT_ID")
	clientSecret = os.Getenv("SANDBOX_CLIENT_SECRET")
	// You have to manually create an auction in the sandbox. Auctions can't be created using the rest api (yet?).
	auctionURL = os.Getenv("SANDOX_AUCTION_URL")
	if clientID == "" || clientSecret == "" || auctionURL == "" {
		panic("Please set SANDBOX_CLIENT_ID, SANDBOX_CLIENT_SECRET and SANDOX_AUCTION_URL.")
	}
}

func TestAuction(t *testing.T) {
	if !integration {
		t.SkipNow()
	}

	ctx := context.Background()

	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://api.sandbox.ebay.com/identity/v1/oauth2/token",
		Scopes:       []string{"https://api.ebay.com/oauth/api_scope"},
	}

	client := ebay.NewSandboxClient(oauth2.NewClient(ctx, ebay.TokenSource(conf.TokenSource(ctx))))

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

	client = ebay.NewSandboxClient(oauth2.NewClient(ctx, ebay.TokenSource(oauthConf.TokenSource(ctx, tok))))

	bid, err := client.Buy.Offer.GetBidding(ctx, it.ItemID, ebay.BuyMarketplaceUSA)
	if err != nil && !ebay.IsError(err, ebay.ErrGetBiddingNoBiddingActivity) {
		t.Fatalf("Expected error code %d, got %+v.", ebay.ErrGetBiddingNoBiddingActivity, err)
	}

	var bidValue, bidCurrency string
	if len(bid.SuggestedBidAmounts) > 0 {
		bidValue = bid.SuggestedBidAmounts[0].Value
		bidCurrency = bid.SuggestedBidAmounts[0].Currency
	} else {
		bidValue = it.CurrentBidPrice.Value
		v, err := strconv.ParseFloat(bidValue, 64)
		if err != nil {
			t.Fatal(err)
		}
		v += 2
		bidValue = fmt.Sprintf("%.2f", v)
		bidCurrency = it.CurrentBidPrice.Currency
	}

	_, err = client.Buy.Offer.PlaceProxyBid(ctx, it.ItemID, ebay.BuyMarketplaceUSA, bidValue, bidCurrency, false)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Successfully bid %+v.", bid.SuggestedBidAmounts[0])
}
