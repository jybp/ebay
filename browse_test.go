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
		fmt.Fprintf(w, `{"itemId": "v1|%s|0"}`, r.URL.Query().Get("legacy_item_id"))
	})

	item, err := client.Buy.Browse.GetItemByLegacyID(context.Background(), "202117468662")
	assert.Nil(t, err)
	assert.Equal(t, "v1|202117468662|0", item.ItemID)
}

func TestGetCompactItem(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/v1|202117468662|0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		fmt.Fprintf(w, `{"itemId": "%s"}`, r.URL.Query().Get("fieldgroups"))
	})

	item, err := client.Buy.Browse.GetCompactItem(context.Background(), "v1|202117468662|0")
	assert.Nil(t, err)
	assert.Equal(t, "COMPACT", item.ItemID)
}

func TestGettItem(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/v1|202117468662|0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		fmt.Fprintf(w, `{"itemId": "%s"}`, r.URL.Query().Get("fieldgroups"))
	})

	item, err := client.Buy.Browse.GetItem(context.Background(), "v1|202117468662|0")
	assert.Nil(t, err)
	assert.Equal(t, "PRODUCT", item.ItemID)
}

func TestGetItemByGroupID(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/get_items_by_item_group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("expected GET method, got: %s", r.Method)
		}
		fmt.Fprintf(w, `{"items": [{"itemId": "%s"}]}`, r.URL.Query().Get("item_group_id"))
	})

	it, err := client.Buy.Browse.GetItemByGroupID(context.Background(), "151915076499")
	assert.Nil(t, err)
	assert.Equal(t, "151915076499", it.Items[0].ItemID)
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
		fmt.Fprintf(w, `{"compatibilityStatus": "NOT_COMPATIBLE", "warnings": [{"category" : "category"}]}`)
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
