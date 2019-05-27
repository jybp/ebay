package ebay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// BrowseService handles communication with the Browse API
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/overview.html
type BrowseService service

// OptBrowseContextualLocation adds the header containing contextualLocation.
// It is strongly recommended that you use it when submitting Browse API methods.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/static/api-browse.html#Headers
func OptBrowseContextualLocation(country, zip string) func(*http.Request) {
	return func(req *http.Request) {
		v := req.Header.Get(headerEndUserCtx)
		if len(v) > 0 {
			v += ","
		}
		v += "contextualLocation=" + url.QueryEscape(fmt.Sprintf("country=%s,zip=%s", country, zip))
		req.Header.Set(headerEndUserCtx, v)
	}
}

// Item represents an eBay item.
type Item struct{}

// GetItem retrieves the details of a specific item.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/browse/resources/item/methods/getItem
func (s *BrowseService) GetItem(ctx context.Context, itemID string, opts ...Opt) (Item, error) {
	u := fmt.Sprintf("buy/browse/v1/item/%s", itemID)
	req, err := s.client.NewRequest(http.MethodGet, u, opts...)
	if err != nil {
		return Item{}, err
	}
	var it Item
	return it, s.client.Do(ctx, req, &it)
}
