package target

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target/anilist"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

//go:generate go run github.com/Khan/genqlient ../../schemas/anilist/genqlient.yaml

type AnilistTarget struct {
	GraphQLTarget
	Target
	client graphql.Client
	ctx    context.Context
	log    *log.Logger
}

func (t *AnilistTarget) Login() (string, error) {
	return "https://anilist.co/api/v2/oauth/authorize?client_id=21288&response_type=token", nil
}

func (t *AnilistTarget) getClient() graphql.Client {
	if t.client == nil {
		t.client = t.GraphQLTarget.getClient(t.ApiUrl, t.GetToken())
	}
	return t.client
}

func (t *AnilistTarget) GetToken() string {
	return config.GetTokens().Anilist
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
	name := "anilist"
	logger := log.WithPrefix(name)
	target := &AnilistTarget{
		ctx: context.Background(),
		log: logger,
		Target: Target{
			Name:   name,
			ApiUrl: "https://graphql.anilist.co",
		},
	}
	return target
}
