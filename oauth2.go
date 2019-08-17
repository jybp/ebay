package ebay

import "golang.org/x/oauth2"

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
