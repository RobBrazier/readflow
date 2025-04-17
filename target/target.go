package target

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/source"
	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-retryablehttp"
)

type Target struct {
	name   string
	ApiUrl string
	SyncTarget
}

func NewTarget(name, url string) Target {
	return Target{
		name:   name,
		ApiUrl: url,
	}
}

type GraphQLTarget struct{}

type authTransport struct {
	key     string
	wrapped http.RoundTripper
}

type SyncTarget interface {
	Login() (string, error)
	GetToken() string
	Name() string
	ShouldProcess(book source.BookContext) bool
	UpdateReadStatus(book source.BookContext) error
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.key))
	return t.wrapped.RoundTrip(req)
}

func (g *GraphQLTarget) GetClient(url, token string) graphql.Client {
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

func (t *Target) Name() string {
	return t.name
}
