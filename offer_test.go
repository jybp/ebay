package ebay_test

import (
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
