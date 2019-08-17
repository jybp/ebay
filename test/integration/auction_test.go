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
	redirectURL  string
)

func init() {
	flag.BoolVar(&integration, "integration", false, "run integration tests")
	flag.Parse()
	if !integration {
		return
	}
	clientID = os.Getenv("SANDBOX_CLIENT_ID")
	clientSecret = os.Getenv("SANDBOX_CLIENT_SECRET")
	redirectURL = os.Getenv("SANDBOX_REDIRECT_URL")
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		panic("Please set SANDBOX_CLIENT_ID, SANDBOX_CLIENT_SECRET and SANDBOX_REDIRECT_URL.")
	}
}

func TestAuction(t *testing.T) {
	if !integration {
		t.SkipNow()
	}

	ctx := context.Background()

	// You have to manually create an auction in the sandbox. Auctions can't be created using the rest api (yet?).
	auctionURL = os.Getenv("SANDOX_AUCTION_URL")

	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     ebay.OAuth20SandboxEndpoint.TokenURL,
		Scopes:       []string{ebay.ScopeRoot},
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
	mux := setupTLS()
	mux.HandleFunc("/accept", func(rw http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/policy", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("eBay Sniper Policy"))
	})
	mux.HandleFunc("/decline", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Decline. You can safely close this tab."))
	})

	oauthConf := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     ebay.OAuth20SandboxEndpoint,
		RedirectURL:  redirectURL,
		Scopes:       []string{ebay.ScopeBuyOfferAuction},
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
