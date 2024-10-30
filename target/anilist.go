package target

import (
	"log/slog"

	"github.com/spf13/viper"
)

type AnilistTarget struct {
	Target
}

func (t *AnilistTarget) Login() (string, error) {
	return "https://anilist.co/api/v2/oauth/authorize?client_id=21288&response_type=token", nil
}

func (t *AnilistTarget) SaveToken(token string) error {
	slog.Info("saved token to", "key", t.getTokenKey())
	viper.Set(t.getTokenKey(), token)
	return nil
}

func NewAnilistTarget() SyncTarget {
	return &AnilistTarget{
		Target: Target{
			Name:     "anilist",
			Hostname: "anilist.co",
			ApiUrl:   "https://graphql.anilist.co",
		},
	}
}

func init() {
	targets = append(targets, NewAnilistTarget())
}
