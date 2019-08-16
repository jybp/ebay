package tokensource

import "golang.org/x/oauth2"

// type TokenSource interface {
// 	// Token returns a token or an error.
// 	// Token must be safe for concurrent use by multiple goroutines.
// 	// The returned Token must not be modified.
// 	Token() (*Token, error)
// }

type TokenSource struct {
	base oauth2.TokenSource
}

func New(base oauth2.TokenSource) *TokenSource {
	return &TokenSource{base: base}
}

func (ts *TokenSource) Token() (*oauth2.Token, error) {
	t, err := ts.base.Token()
	if t != nil {
		print("new token: " + t.AccessToken + "\n") // TODO remove
		t.TokenType = "Bearer"
	}
	return t, err
}
