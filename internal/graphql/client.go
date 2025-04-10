package graphql

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal"
	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-retryablehttp"
)

type authTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.key))
	req.Header.Set("User-Agent", fmt.Sprintf("readflow/%s (https://github.com/RobBrazier/readflow)", internal.Version))
	return t.wrapped.RoundTrip(req)
}

func GetClient(url, token string) graphql.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{
		Transport: &authTransport{
			key:     token,
			wrapped: http.DefaultTransport,
		},
	}
	retryClient.Logger = slog.New(log.WithPrefix("graphql"))
	httpClient := retryClient.StandardClient()
	return graphql.NewClient(url, httpClient)
}
