package ebay_test

import (
	"context"
	"fmt"
	"net/http"
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
		marketplaceID := r.Header.Get("X-EBAY-C-MARKETPLACE-ID")
		fmt.Fprintf(w, `{"itemId": "%s"}`, marketplaceID)
	})

	bidding, err := client.Buy.Offer.GetBidding(context.Background(), "v1|202117468662|0", ebay.BuyMarketplaceUSA)
	assert.Nil(t, err)
	assert.Equal(t, ebay.BuyMarketplaceUSA, bidding.ItemID)
}
