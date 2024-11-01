package target

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Target struct {
	Name     string
	Hostname string
	ApiUrl   string
	SyncTarget
}

type GraphQLTarget struct{}

type authTransport struct {
	key     string
	wrapped http.RoundTripper
}

type SyncTarget interface {
	Login() (string, error)
	HasToken() bool
	GetTarget() *Target
	getToken() string
	getTokenKey() string
	SaveToken(token string) error
	GetName() string
	GetHostname() string
	GetCurrentUser() string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.key))
	return t.wrapped.RoundTrip(req)
}

func (g *GraphQLTarget) getClient(target Target) graphql.Client {
	httpClient := http.Client{
		Transport: &authTransport{
			key:     target.getToken(),
			wrapped: http.DefaultTransport,
		},
	}
	return graphql.NewClient(target.ApiUrl, &httpClient)
}

var targets = []SyncTarget{}

func (t *Target) GetTarget() *Target {
	return t
}

func (t *Target) GetName() string {
	return t.Name
}

func (t *Target) GetHostname() string {
	return t.Hostname
}

func (t *Target) getToken() string {
	key := t.getTokenKey()
	value := viper.GetString(key)
	if value == "" {
		cobra.CheckErr(fmt.Sprintf("Token for %s not set - please configure with `%s login` or disable with `%s config set targets DESIRED_TARGET`", t.Name, internal.NAME, internal.NAME))
	}
	return value
}

func (t *Target) getTokenKey() string {
	return fmt.Sprintf("tokens.%s", t.Name)
}

func (t *Target) HasToken() bool {
	return t.getToken() != ""
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
