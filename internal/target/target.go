package target

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/viper"
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
	getToken() string
	GetName() string
	GetCurrentUser() string
	UpdateReadStatus(book source.BookContext) error
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.key))
	return t.wrapped.RoundTrip(req)
}

func (g *GraphQLTarget) getClient(target Target) graphql.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{
		Transport: &authTransport{
			key:     target.getToken(),
			wrapped: http.DefaultTransport,
		},
	}
	retryClient.Logger = slog.Default()
	httpClient := retryClient.StandardClient()
	return graphql.NewClient(target.ApiUrl, httpClient)
}

var targets = []SyncTarget{}

func (t *Target) GetName() string {
	return t.Name
}

func (t *Target) getToken() string {
	value := viper.GetString(fmt.Sprintf("tokens.%s", t.Name))
	return value
}

func GetTargets() []SyncTarget {
	if len(targets) == 0 {
		targets = []SyncTarget{
			NewAnilistTarget(),
			NewHardcoverTarget(),
		}
	}
	return targets
}

func GetActiveTargets() []SyncTarget {
	active := []SyncTarget{}
	selectedTargets := viper.GetStringSlice("targets")
	for _, target := range GetTargets() {
		if slices.Contains(selectedTargets, target.GetName()) {
			active = append(active, target)
		}
	}
	return active
}
