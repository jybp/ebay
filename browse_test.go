package ebay_test

import (
	"net/http"
	"testing"

	"github.com/jybp/ebay"
	"github.com/stretchr/testify/assert"
)

func TestOptContextualLocation(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	ebay.OptContextualLocation("US", "19406")(r)
	assert.Equal(t, "country%3DUS%2Czip%3D19406", r.Header.Get("X-EBAY-C-ENDUSERCTX"))
}
