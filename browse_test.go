package ebay_test

import (
	"context"
	"fmt"
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
		fmt.Fprintf(w, `{"itemId": "%s"}`, r.URL.Query().Get("fieldgroups"))
	})

	item, err := client.Buy.Browse.GetItem(context.Background(), "v1|202117468662|0")
	assert.Nil(t, err)
	assert.Equal(t, "PRODUCT", item.ItemID)
}
