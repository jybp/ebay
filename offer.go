package ebay

import (
	"net/http"
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

// OptBuyMarketplace adds the header containing the marketplace id:
// https://developer.ebay.com/api-docs/buy/static/ref-marketplace-supported.html
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/static/api-browse.html#Headers
func OptBuyMarketplace(marketplaceID string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("X-EBAY-C-MARKETPLACE-ID", marketplaceID)
	}
}
