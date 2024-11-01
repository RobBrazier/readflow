package target

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/target/hardcover"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:generate go run github.com/Khan/genqlient ../schemas/hardcover/genqlient.yaml

type HardcoverTarget struct {
	Target
	GraphQLTarget
	ctx    context.Context
	client graphql.Client
}

func (t *HardcoverTarget) Login() (string, error) {
	return "https://hardcover.app/account/api", nil
}

func (t *HardcoverTarget) getClient() graphql.Client {
	if t.client == nil {
		t.client = t.GraphQLTarget.getClient(t.Target)
	}
	return t.client
}

func (t *HardcoverTarget) SaveToken(token string) error {
	slog.Info("saved token to", "key", t.getTokenKey())
	token = strings.TrimSpace(strings.Replace(token, "Bearer", "", 1))
	viper.Set(t.getTokenKey(), token)
	return nil
}

func (t *HardcoverTarget) GetCurrentUser() string {
	response, err := hardcover.GetCurrentUser(t.ctx, t.getClient())
	cobra.CheckErr(err)
	return response.GetMe()[0].GetUsername()
}

func NewHardcoverTarget() SyncTarget {
	target := &HardcoverTarget{
		ctx: context.Background(),
		Target: Target{
			Name:     "hardcover",
			Hostname: "hardcover.app",
			ApiUrl:   "https://api.hardcover.app/v1/graphql",
		},
	}
	return target
}
