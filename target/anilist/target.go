package anilist

import (
	"context"

	"github.com/RobBrazier/readflow/internal"
)

type anilistTarget struct {
	internal.Target
}

func (t anilistTarget) Login() (string, error) {
	return "https://anilist.co/api/v2/oauth/authorize?client_id=21288&response_type=token", nil
}

func New(ctx context.Context) internal.SyncTarget {
	return nil
}
