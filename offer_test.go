package ebay_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/jybp/ebay"
	"github.com/stretchr/testify/assert"
)

func TestOptBuyMarketplace(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	ebay.OptBuyMarketplace("EBAY_US")(r)
	assert.Equal(t, "EBAY_US", r.Header.Get("X-EBAY-C-MARKETPLACE-ID"))
}

func TestGetBidding(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/offer/v1_beta/bidding/v1|202117468662|0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		marketplaceID := r.Header.Get("X-EBAY-C-MARKETPLACE-ID")
		fmt.Fprintf(w, `{"itemId": "%s"}`, marketplaceID)
	})

	bidding, err := client.Buy.Offer.GetBidding(context.Background(), "v1|202117468662|0", ebay.BuyMarketplaceUSA)
	assert.Nil(t, err)
	assert.Equal(t, ebay.BuyMarketplaceUSA, bidding.ItemID)
}

func TestPlaceProxyBid(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/offer/v1_beta/bidding/v1|202117468662|0/place_proxy_bid", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST method, got: %s", r.Method)
		}
		marketplaceID := r.Header.Get("X-EBAY-C-MARKETPLACE-ID")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("%+v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		escapedBody := strconv.Quote(string(body))
		escapedBody = escapedBody[1 : len(escapedBody)-1]
		fmt.Fprintf(w, `{"proxyBidId": "%s_%s"}`, escapedBody, marketplaceID)
	})

	bid, err := client.Buy.Offer.PlaceProxyBid(context.Background(), "v1|202117468662|0", ebay.BuyMarketplaceUSA, "1.23", "USD", false)
	assert.Nil(t, err)
	assert.Equal(t, "{\"maxAmount\":{\"currency\":\"USD\",\"value\":\"1.23\"}}\n_EBAY_US", bid.ProxyBidID)

	bid, err = client.Buy.Offer.PlaceProxyBid(context.Background(), "v1|202117468662|0", ebay.BuyMarketplaceUSA, "1.23", "USD", true)
	assert.Nil(t, err)
	assert.Equal(t, "{\"maxAmount\":{\"currency\":\"USD\",\"value\":\"1.23\"},\"userConsent\":{\"adultOnlyItem\":true}}\n_EBAY_US", bid.ProxyBidID)
}
