package ebay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	baseURL        = "https://api.ebay.com/"
	sandboxBaseURL = "https://api.sandbox.ebay.com/"

	headerEndUserCtx = "X-EBAY-C-ENDUSERCTX"
)

// BuyAPI regroups the eBay Buy APIs.
//
// eBay API docs: https://developer.ebay.com/api-docs/buy/static/buy-landing.html
type BuyAPI struct {
	Browse *BrowseService
}

// Client manages communication with the eBay API.
type Client struct {
	client  *http.Client // Used to make actual API requests.
	baseURL *url.URL     // Base URL for API requests.

	// eBay APIs.
	Buy BuyAPI
}

// NewClient returns a new eBay API client.
// If a nil httpClient is provided, http.DefaultClient will be used.
func NewClient(httpclient *http.Client) *Client {
	return newClient(httpclient, baseURL)
}

// NewSandboxClient returns a new eBay sandbox API client.
// If a nil httpClient is provided, http.DefaultClient will be used.
func NewSandboxClient(httpclient *http.Client) *Client {
	return newClient(httpclient, sandboxBaseURL)
}

// NewCustomClient returns a new custom eBay API client.
// BaseURL should have a trailing slash.
// If a nil httpClient is provided, http.DefaultClient will be used.
func NewCustomClient(httpclient *http.Client, baseURL string) (*Client, error) {
	if !strings.HasSuffix(baseURL, "/") {
		return nil, fmt.Errorf("BaseURL %s must have a trailing slash", baseURL)
	}
	return newClient(httpclient, baseURL), nil
}

func newClient(httpclient *http.Client, baseURL string) *Client {
	if httpclient == nil {
		httpclient = http.DefaultClient
	}
	url, _ := url.Parse(baseURL)
	c := &Client{client: httpclient, baseURL: url}
	c.Buy = BuyAPI{
		Browse: (*BrowseService)(&service{c}),
	}
	return c
}

type service struct {
	client *Client
}

// Opt describes functional options for the eBay API.
type Opt func(*http.Request)

// NewRequest creates an API request.
// url should always be specified without a preceding slash.
func (c *Client) NewRequest(method, url string, opts ...Opt) (*http.Request, error) {
	if strings.HasPrefix(url, "/") {
		return nil, errors.New("url should always be specified without a preceding slash")
	}
	u, err := c.baseURL.Parse(url)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(req)
	}
	return req, nil
}

// Do sends an API request and stores the JSON decoded value into v.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := CheckResponse(resp); err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// An ErrorData reports one or more errors caused by an API request.
//
// eBay API docs: https://developer.ebay.com/api-docs/static/handling-error-messages.html
type ErrorData struct {
	Errors []struct {
		ErrorID     int      `json:"errorId,omitempty"`
		Domain      string   `json:"domain,omitempty"`
		SubDomain   string   `json:"subDomain,omitempty"`
		Category    string   `json:"category,omitempty"`
		Message     string   `json:"message,omitempty"`
		LongMessage string   `json:"longMessage,omitempty"`
		InputRefIds []string `json:"inputRefIds,omitempty"`
		OuputRefIds []string `json:"outputRefIds,omitempty"`
		Parameters  []struct {
			Name  string `json:"name,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"parameters,omitempty"`
	} `json:"errors,omitempty"`
	Response *http.Response
}

func (e *ErrorData) Error() string {
	return fmt.Sprintf("%s %s: %d %+v", e.Response.Request.Method, e.Response.Request.URL,
		e.Response.StatusCode, e.Errors)
}

// CheckResponse checks the API response for errors, and returns them if present.
func CheckResponse(resp *http.Response) error {
	if s := resp.StatusCode; 200 <= s && s < 300 {
		return nil
	}
	errorData := &ErrorData{Response: resp}
	_ = json.NewDecoder(resp.Body).Decode(errorData)
	return errorData
}
