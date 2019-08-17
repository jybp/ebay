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

// Some valid eBay error codes for the GetBidding method.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/offer/resources/bidding/methods/getBidding#h2-error-codes
const (
	ErrGetBiddingMarketplaceNotSupported = 120017
	ErrGetBiddingNoBiddingActivity       = 120033
)

// GetBidding retrieves the buyer's bidding details on a specific auction item.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/offer/resources/bidding/methods/getBidding
func (s *OfferService) GetBidding(ctx context.Context, itemID, marketplaceID string, opts ...Opt) (Bidding, error) {
	u := fmt.Sprintf("buy/offer/v1_beta/bidding/%s", itemID)
	opts = append(opts, OptBuyMarketplace(marketplaceID))
	req, err := s.client.NewRequest(http.MethodGet, u, nil, opts...)
	if err != nil {
		return Bidding{}, err
	}
	var bid Bidding
	return bid, s.client.Do(ctx, req, &bid)
}

// ProxyBid represents an eBay proxy bid.
type ProxyBid struct {
	ProxyBidID string `json:"proxyBidId"`
}

// Some valid eBay error codes for the PlaceProxyBid method.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/offer/resources/bidding/methods/placeProxyBid#h2-error-codes
const (
	ErrPlaceProxyBidAuctionEndedBecauseOfBuyItNow       = 120002
	ErrPlaceProxyBidBidCannotBeGreaterThanBuyItNowPrice = 120005
	ErrPlaceProxyBidAmountTooHigh                       = 120007
	ErrPlaceProxyBidAmountTooLow                        = 120008
	ErrPlaceProxyBidCurrencyMustMatchItemPriceCurrency  = 120009
	ErrPlaceProxyBidCannotLowerYourProxyBid             = 120010
	ErrPlaceProxyBidAmountExceedsLimit                  = 120011
	ErrPlaceProxyBidAuctionHasEnded                     = 120012
	ErrPlaceProxyBidAmountInvalid                       = 120013
	ErrPlaceProxyBidCurrencyInvalid                     = 120014
	ErrPlaceProxyBidMaximumBidAmountMissing             = 120016
)

// PlaceProxyBid places a proxy bid for the buyer on a specific auction item.
//
// Curency is the three-letter ISO 4217 code representing the currency.
// For one hundred US dollars, MaxAmout is "100.00" and currency is "USD".
//
// You must ensure the user agrees to the "Terms of use for Adult Only category"
// (https://signin.ebay.com/ws/eBayISAPI.dll?AdultSignIn2) if he wishes to bid on on a adult-only item.
// An item is adult-only if the AdultOnly field returned by the Browse API is set to true.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/offer/resources/bidding/methods/placeProxyBid
func (s *OfferService) PlaceProxyBid(ctx context.Context, itemID, marketplaceID, maxAmount, currency string, userConsentAdultOnlyItem bool, opts ...Opt) (ProxyBid, error) {
	u := fmt.Sprintf("buy/offer/v1_beta/bidding/%s/place_proxy_bid", itemID)
	opts = append(opts, OptBuyMarketplace(marketplaceID))
	type userConsent struct {
		AdultOnlyItem bool `json:"adultOnlyItem,omitempty"`
	}
	type amount struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	}
	type payload struct {
		MaxAmount   amount       `json:"maxAmount"`
		UserConsent *userConsent `json:"userConsent,omitempty"`
	}
	pl := payload{
		MaxAmount: amount{Currency: currency, Value: maxAmount},
	}
	if userConsentAdultOnlyItem {
		pl.UserConsent = &userConsent{userConsentAdultOnlyItem}
	}
	req, err := s.client.NewRequest(http.MethodPost, u, &pl, opts...)
	if err != nil {
		return ProxyBid{}, err
	}
	var bid ProxyBid
	return bid, s.client.Do(ctx, req, &bid)
}
