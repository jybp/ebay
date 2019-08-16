package ebay

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// OfferService handles communication with the Offer API
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/offer/static/overview.html
type OfferService service

// Valid marketplace IDs
const (
	BuyMarketplaceAustralia    = "EBAY_AU"
	BuyMarketplaceCanada       = "EBAY_CA"
	BuyMarketplaceGermany      = "EBAY_DE"
	BuyMarketplaceSpain        = "EBAY_ES"
	BuyMarketplaceFrance       = "EBAY_FR"
	BuyMarketplaceGreatBritain = "EBAY_GB"
	BuyMarketplaceHongKong     = "EBAY_HK"
	BuyMarketplaceItalia       = "EBAY_IT"
	BuyMarketplaceUSA          = "EBAY_US"
)

// Valid values for the "auctionStatus" Bidding field.
const (
	BiddingAuctionStatusEnded = "ENDED"
)

// OptBuyMarketplace adds the header containing the marketplace id:
// https://developer.ebay.com/api-docs/buy/static/ref-marketplace-supported.html
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/static/api-browse.html#Headers
func OptBuyMarketplace(marketplaceID string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("X-EBAY-C-MARKETPLACE-ID", marketplaceID)
	}
}

// Bidding represents an eBay item bidding.
type Bidding struct {
	AuctionStatus  string    `json:"auctionStatus"`
	AuctionEndDate time.Time `json:"auctionEndDate"`
	ItemID         string    `json:"itemId"`
	CurrentPrice   struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"currentPrice"`
	BidCount            int  `json:"bidCount"`
	HighBidder          bool `json:"highBidder"`
	ReservePriceMet     bool `json:"reservePriceMet"`
	SuggestedBidAmounts []struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"suggestedBidAmounts"`
	CurrentProxyBid struct {
		ProxyBidID string `json:"proxyBidId"`
		MaxAmount  struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"maxAmount"`
	} `json:"currentProxyBid"`
}

// GetBidding retrieves the buyer's bidding details on an auction.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/offer/resources/bidding/methods/getBidding
func (s *OfferService) GetBidding(ctx context.Context, itemID, marketplaceID string, opts ...Opt) (Item, error) {
	u := fmt.Sprintf("buy/offer/v1_beta/bidding/%s", itemID)
	opts = append(opts, OptBuyMarketplace(marketplaceID))
	req, err := s.client.NewRequest(http.MethodGet, u, opts...)
	if err != nil {
		return Item{}, err
	}
	var it Item
	return it, s.client.Do(ctx, req, &it)
}
