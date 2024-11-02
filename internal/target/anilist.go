package target

import (
	"context"
	"log/slog"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target/anilist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:generate go run github.com/Khan/genqlient ../schemas/anilist/genqlient.yaml

type AnilistTarget struct {
	GraphQLTarget
	Target
	client graphql.Client
	ctx    context.Context
}

func (t *AnilistTarget) Login() (string, error) {
	return "https://anilist.co/api/v2/oauth/authorize?client_id=21288&response_type=token", nil
}

func (t *AnilistTarget) getClient() graphql.Client {
	if t.client == nil {
		t.client = t.GraphQLTarget.getClient(t.Target)
	}
	return t.client
}

func (t *AnilistTarget) SaveToken(token string) error {
	slog.Info("saved token to", "key", t.getTokenKey())
	viper.Set(t.getTokenKey(), token)
	return nil
}

func (t *AnilistTarget) GetCurrentUser() string {
	response, err := anilist.GetCurrentUser(t.ctx, t.getClient())
	cobra.CheckErr(err)
	return response.Viewer.Name
}

func (t *AnilistTarget) UpdateReadStatus(book source.BookContext) error {
	return nil
}

func NewAnilistTarget() SyncTarget {
	target := &AnilistTarget{
		ctx: context.Background(),
		Target: Target{
			Name:     "anilist",
			Hostname: "anilist.co",
			ApiUrl:   "https://graphql.anilist.co",
		},
	}
	return target
}
