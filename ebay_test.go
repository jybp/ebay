package ebay_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jybp/ebay"
	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	testOpt := func(r *http.Request) {
		r.URL.RawQuery = "q=1"
	}
	client, _ := ebay.NewCustomClient(nil, "https://api.ebay.com/")
	r, _ := client.NewRequest(http.MethodPost, "test", testOpt)
	assert.Equal(t, "https://api.ebay.com/test?q=1", fmt.Sprint(r.URL))
	assert.Equal(t, http.MethodPost, r.Method)
}

func TestCheckResponseNoError(t *testing.T) {
	resp := &http.Response{StatusCode: 200}
	assert.Nil(t, ebay.CheckResponse(resp))
}

func TestCheckResponse(t *testing.T) {
	body := ` {
		"errors": [
			{
				"errorId": 15008,
				"domain": "API_ORDER",
				"subDomain": "subdomain",
				"category": "REQUEST",
				"message": "Invalid Field : itemId.",
				"longMessage": "longMessage",
				"inputRefIds": [
					"$.lineItemInputs[0].itemId"
				],
				"outputRefIds": [
					"outputRefId"
				],
				"parameters": [
					{
					"name": "itemId",
					"value": "2200077988|0"
					}
				]
			}
		]
	}`
	resp := &http.Response{StatusCode: 400, Body: ioutil.NopCloser(bytes.NewBufferString(body))}
	err, ok := ebay.CheckResponse(resp).(*ebay.ErrorData)
	assert.True(t, ok)
	assert.Equal(t, 1, len(err.Errors))
	assert.Equal(t, 15008, err.Errors[0].ErrorID)
	assert.Equal(t, "API_ORDER", err.Errors[0].Domain)
	assert.Equal(t, "subdomain", err.Errors[0].SubDomain)
	assert.Equal(t, "REQUEST", err.Errors[0].Category)
	assert.Equal(t, "Invalid Field : itemId.", err.Errors[0].Message)
	assert.Equal(t, "longMessage", err.Errors[0].LongMessage)
	assert.Equal(t, []string{"$.lineItemInputs[0].itemId"}, err.Errors[0].InputRefIds)
	assert.Equal(t, []string{"outputRefId"}, err.Errors[0].OuputRefIds)
	assert.Equal(t, "itemId", err.Errors[0].Parameters[0].Name)
	assert.Equal(t, "2200077988|0", err.Errors[0].Parameters[0].Value)
}
