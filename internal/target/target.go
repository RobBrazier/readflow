package target

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-retryablehttp"
)

var (
	targets     atomic.Pointer[[]SyncTarget]
	targetsOnce sync.Once
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
	retryClient.Logger = slog.New(log.Default())
	httpClient := retryClient.StandardClient()
	return graphql.NewClient(url, httpClient)
}

func (t *Target) GetName() string {
	return t.Name
}

func GetTargets() *[]SyncTarget {
	t := targets.Load()
	if t == nil {
		targetsOnce.Do(func() {
			targets.CompareAndSwap(nil, &[]SyncTarget{
				NewAnilistTarget(),
				NewHardcoverTarget(),
			})
		})
		t = targets.Load()
	}
	return t
}

func GetActiveTargets() []SyncTarget {
	active := []SyncTarget{}
	selectedTargets := config.GetTargets()
	for _, target := range *GetTargets() {
		if slices.Contains(selectedTargets, target.GetName()) {
			active = append(active, target)
		}
	}
	return active
}
