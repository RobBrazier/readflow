package target

import (
	"fmt"
	"slices"

	"github.com/spf13/viper"
)

type Target struct {
	Name     string
	Hostname string
	ApiUrl   string
	SyncTarget
}

var targets = []SyncTarget{}

type baseTarget struct {
	Target Target
	SyncTarget
}

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
	return viper.GetString(t.getTokenKey())
}

func (t *Target) getTokenKey() string {
	return fmt.Sprintf("tokens.%s", t.Name)
}

func (t *Target) HasToken() bool {
	return t.getToken() != ""
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
}

func GetTargets() []SyncTarget {
	return targets
}

func GetActiveTargets() []SyncTarget {
	active := []SyncTarget{}
	selectedTargets := viper.GetStringSlice("targets")
	for _, target := range targets {
		if slices.Contains(selectedTargets, target.GetName()) {
			active = append(active, target)
		}
	}
	return active
}
