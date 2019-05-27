package ebay_test

import (
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
