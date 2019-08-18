package ebay_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jybp/ebay"
	"github.com/stretchr/testify/assert"
)

func TestOptBrowseContextualLocation(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	ebay.OptBrowseContextualLocation("US", "19406")(r)
	assert.Equal(t, "contextualLocation=country%3DUS%2Czip%3D19406", r.Header.Get("X-EBAY-C-ENDUSERCTX"))
}

func TestOptBrowseContextualLocationExistingHeader(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	r.Header.Set("X-EBAY-C-ENDUSERCTX", "affiliateCampaignId=1")
	ebay.OptBrowseContextualLocation("US", "19406")(r)
	assert.Equal(t, "affiliateCampaignId=1,contextualLocation=country%3DUS%2Czip%3D19406", r.Header.Get("X-EBAY-C-ENDUSERCTX"))
}

func TestGetLegacyItem(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/get_item_by_legacy_id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		assert.Equal(t, "202117468662", r.URL.Query().Get("legacy_item_id"))
		fmt.Fprintf(w, `{"itemId": "itemId"}`)
	})

	item, err := client.Buy.Browse.GetItemByLegacyID(context.Background(), "202117468662")
	assert.Nil(t, err)
	assert.Equal(t, "itemId", item.ItemID)
}

func TestGetCompactItem(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/v1|202117468662|0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		assert.Equal(t, "COMPACT", r.URL.Query().Get("fieldgroups"))
		fmt.Fprintf(w, `{"itemId": "itemId"}`)
	})

	item, err := client.Buy.Browse.GetCompactItem(context.Background(), "v1|202117468662|0")
	assert.Nil(t, err)
	assert.Equal(t, "itemId", item.ItemID)
}

func TestGettItem(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/v1|202117468662|0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		assert.Equal(t, "PRODUCT", r.URL.Query().Get("fieldgroups"))
		fmt.Fprint(w, `{"itemId": "itemId"}`)
	})

	item, err := client.Buy.Browse.GetItem(context.Background(), "v1|202117468662|0")
	assert.Nil(t, err)
	assert.Equal(t, "itemId", item.ItemID)
}

func TestGetItemByGroupID(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/get_items_by_item_group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		assert.Equal(t, "151915076499", r.URL.Query().Get("item_group_id"))
		fmt.Fprint(w, `{"items": [{"itemId": "itemId"}]}`)
	})

	it, err := client.Buy.Browse.GetItemByGroupID(context.Background(), "151915076499")
	assert.Nil(t, err)
	assert.Equal(t, "itemId", it.Items[0].ItemID)
}

func TestCheckCompatibility(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/v1|202117468662|0/check_compatibility", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST method, got: %s", r.Method)
		}
		assert.Equal(t, ebay.BuyMarketplaceUSA, r.Header.Get("X-EBAY-C-MARKETPLACE-ID"))
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		assert.Equal(t, `{"compatibilityProperties":[{"name":"0","value":"1"},{"name":"2","value":"3"}]}
`, string(body))
		fmt.Fprint(w, `{"compatibilityStatus": "NOT_COMPATIBLE", "warnings": [{"category" : "category"}]}`)
	})
	compatibilityProperties := []ebay.CompatibilityProperty{
		{Name: "0", Value: "1"},
		{Name: "2", Value: "3"},
	}
	compatibility, err := client.Buy.Browse.CheckCompatibility(context.Background(), "v1|202117468662|0", ebay.BuyMarketplaceUSA, compatibilityProperties)
	assert.Nil(t, err)
	assert.Equal(t, "NOT_COMPATIBLE", compatibility.CompatibilityStatus)
	assert.Equal(t, "category", compatibility.Warnings[0].Category)
}

func TestSearch(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item_summary/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		assert.Equal(t, "search", r.URL.Query().Get("q"))
		assert.Equal(t, "2", r.URL.Query().Get("limit"))

		fmt.Fprint(w, `{"href": "href","total":1,"itemSummaries": [{"itemId": "itemId"}]}`)
	})

	search, err := client.Buy.Browse.Search(context.Background(), ebay.OptBrowseSearch("search"), ebay.OptBrowseSearchLimit(2))
	assert.Nil(t, err)
	assert.Equal(t, "href", search.Href)
	assert.Equal(t, 1, search.Total)
	assert.Equal(t, "itemId", search.ItemSummaries[0].ItemID)
}
