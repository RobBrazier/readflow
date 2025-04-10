package hardcover

import (
	"context"
	"sync"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/config"
	gql "github.com/RobBrazier/readflow/internal/graphql"
)

var lock = &sync.Mutex{}

var client graphql.Client

func GetClient(_ context.Context) (graphql.Client, error) {
	if client == nil {
		lock.Lock()
		defer lock.Unlock()
		url := config.GetApiConfig().HardcoverEndpoint
		token := config.GetTokens().Hardcover
		client = gql.GetClient(url, token)
	}
	return client, nil
}
