package target

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"

	slogzerolog "github.com/samber/slog-zerolog/v2"
)

type Target struct {
	Name   string
	ApiUrl string
	SyncTarget
}

type GraphQLTarget struct{}

type authTransport struct {
	key     string
	wrapped http.RoundTripper
}

type SyncTarget interface {
	Login() (string, error)
	GetToken() string
	GetName() string
	GetCurrentUser() string
	ShouldProcess(book source.BookContext) bool
	UpdateReadStatus(book source.BookContext) error
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.key))
	return t.wrapped.RoundTrip(req)
}

func (g *GraphQLTarget) getClient(url, token string) graphql.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{
		Transport: &authTransport{
			key:     token,
			wrapped: http.DefaultTransport,
		},
	}
	logger := log.With().Str("client", "graphql").Logger()
	retryClient.Logger = slog.New(
		slogzerolog.Option{
			Level:  slog.LevelDebug,
			Logger: &logger,
		}.NewZerologHandler(),
	)
	httpClient := retryClient.StandardClient()
	return graphql.NewClient(url, httpClient)
}

func (t *Target) GetName() string {
	return t.Name
}

func GetTargets(ctx context.Context) map[string]SyncTarget {
	return map[string]SyncTarget{
		"anilist":   NewAnilistTarget(ctx),
		"hardcover": NewHardcoverTarget(ctx),
	}
}
