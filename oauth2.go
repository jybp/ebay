package ebay

import "golang.org/x/oauth2"

// eBay OAuth 2.0 endpoints.
var (
	OAuth20Endpoint = oauth2.Endpoint{
		AuthURL:  "https://auth.ebay.com/oauth2/authorize",
		TokenURL: "https://api.ebay.com/identity/v1/oauth2/token",
	}
	OAuth20SandboxEndpoint = oauth2.Endpoint{
		AuthURL:  "https://auth.sandbox.ebay.com/oauth2/authorize",
		TokenURL: "https://api.sandbox.ebay.com/identity/v1/oauth2/token",
	}
)

// BearerTokenSource forces the type of the token returned by the 'base' TokenSource to 'Bearer'.
// The eBay API will return "Application Access Token" or "User Access Token" as token_type but
// it must be set to 'Bearer' for subsequent requests.
type BearerTokenSource struct {
	base oauth2.TokenSource
}

// TokenSource returns a new BearerTokenSource.
func TokenSource(base oauth2.TokenSource) *BearerTokenSource {
	return &BearerTokenSource{base: base}
}

// Token allows BearerTokenSource to implement oauth2.TokenSource.
func (ts *BearerTokenSource) Token() (*oauth2.Token, error) {
	t, err := ts.base.Token()
	if t != nil {
		t.TokenType = "Bearer"
	}
	return t, err
}
