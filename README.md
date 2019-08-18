# ebay

ebay is a Go client library for accessing the [eBay API](https://developer.ebay.com/).

## Usage

Create a new eBay API client using the New... functions to access the eBay API:

```go
client := ebay.NewClient(nil)
// Search for iphones or ipads sorted by price in ascending order.
search, err := client.Buy.Browse.Search(context.Background(), ebay.OptBrowseSearch("iphone ipad"), ebay.OptBrowseSearchSort("price"))
```

## Authentication

The ebay library does not directly handle authentication. Instead, when creating a new client, pass an http.Client that can handle authentication for you.

A `TokenSource` function is provided if you are using `golang.org/x/oauth2`. It overrides the token type to `Bearer` so your requests won't fail.

An example for the [client credentials grant flow](https://developer.ebay.com/api-docs/static/oauth-client-credentials-grant.html):

```go
import (
    "context"
    "golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	cfg := clientcredentials.Config{
		ClientID:     "your client id",
		ClientSecret: "your client secret",
		TokenURL:     ebay.OAuth20SandboxEndpoint.TokenURL,
		Scopes:       []string{ebay.ScopeRoot /* your scopes */},
    }
    tc := oauth2.NewClient(context.Background(), ebay.TokenSource(cfg.TokenSource(ctx)))
    client := ebay.NewSandboxClient(tc)

    // Get an item detail.
    result, err := client.Buy.Browse.GetItem(context.Background(), "v1|123456789012|0")
}
```

An example for the [authorization code grant flow](https://developer.ebay.com/api-docs/static/oauth-authorization-code-grant.html):


```go
import (
    "context"
    "golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
    cfg := oauth2.Config{
		ClientID:     "your client id",
		ClientSecret: "your client id",
		Endpoint:     ebay.OAuth20SandboxEndpoint,
		RedirectURL:  "your eBay Redirect URL name (RuName)",
		Scopes:       []string{ebay.ScopeBuyOfferAuction /* your scopes */},
	}

	url := cfg.AuthCodeURL(state)
    fmt.Printf("Visit the URL: %v\n", url)
    
	var authCode string /* Retrieve the authorization code. */

	tok, err := oauthConf.Exchange(ctx, authCode)
	if err != nil {
		panic(err)
	}
	client := ebay.NewSandboxClient(oauth2.NewClient(ctx, ebay.TokenSource(cfg.TokenSource(ctx, tok))))

    // Get bidding for the authenticated user.
    bidding, err := client.Buy.Offer.GetBidding(ctx, "v1|123456789012|0", ebay.BuyMarketplaceUSA)
}
```

## Support

Currently, only some Buy APIs are supported:

| API | Resource |
| --- | --- | 
| Browse | item_summary |
| Browse | item |
| Offer | bidding |


## Documentation

https://godoc.org/github.com/jybp/ebay

## Thanks

- https://github.com/mholt/json-to-go for saving everyone's time.
- https://github.com/google/go-github for their library design.