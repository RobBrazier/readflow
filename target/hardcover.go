package target

import (
	"log/slog"

	"github.com/spf13/viper"
)

type HardcoverTarget struct {
	Target
}

func (t *HardcoverTarget) Login() (string, error) {
	return "https://hardcover.app/account/api", nil
}

func (t *HardcoverTarget) SaveToken(token string) error {
	slog.Info("saved token to", "key", t.getTokenKey())
	viper.Set(t.getTokenKey(), token)
	return nil
}

func NewHardcoverTarget() SyncTarget {
	return &HardcoverTarget{
		Target: Target{
			Name:     "hardcover",
			Hostname: "hardcover.app",
			ApiUrl:   "https://api.hardcover.app/v1/graphql",
		},
	}
}

func init() {
	targets = append(targets, NewHardcoverTarget())
}
