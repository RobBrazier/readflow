package target

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-retryablehttp"
)

type Target struct {
	name   string
	apiUrl string
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
	Name() string
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
	retryClient.Logger = slog.New(log.WithPrefix("graphql"))
	httpClient := retryClient.StandardClient()
	return graphql.NewClient(url, httpClient)
}

func (t *Target) Name() string {
	return t.name
}

type SyncTargetFunc func(ctx context.Context) SyncTarget

func TargetProvider(fn SyncTargetFunc) SyncTargetFunc {
	return func(ctx context.Context) SyncTarget {
		return fn(ctx)
	}
}

func GetTargets() map[string]SyncTargetFunc {
	return map[string]SyncTargetFunc{
		"anilist":   TargetProvider(NewAnilistTarget),
		"hardcover": TargetProvider(NewHardcoverTarget),
	}
}

func GetActiveTargets(enabled []string, ctx context.Context) (active []SyncTarget) {
	for name, target := range GetTargets() {
		if slices.Contains(enabled, name) {
			active = append(active, target(ctx))
		}
	}
	return active
}
