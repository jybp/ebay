package ebay_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/jybp/ebay"
	"github.com/stretchr/testify/assert"
)

func TestOptBrowseContextualLocationn(t *testing.T) {
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

func TestGetItem(t *testing.T) {
	client, mux, teardown := setup(t)
	defer teardown()

	mux.HandleFunc("/buy/browse/v1/item/v1|202117468662|0", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"itemId": "v1|202117468662|0", "title": "%s"}`, r.URL.Query().Get("fieldgroups"))
	})

	item, err := client.Buy.Browse.GetItem(context.Background(), "v1|202117468662|0", "COMPACT")
	assert.Nil(t, err)
	assert.Equal(t, "v1|202117468662|0", item.ItemID)
	assert.Equal(t, "COMPACT", item.Title)
}
