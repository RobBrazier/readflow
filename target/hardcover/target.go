package hardcover

import (
	"context"
	"fmt"
	"strconv"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/model"
	"github.com/charmbracelet/log"
)

type hardcoverTarget struct {
	ctx context.Context
	internal.Target
}

func (t hardcoverTarget) Login() (string, error) {
	return "https://hardcover.app/account/api", nil
}

func (t hardcoverTarget) getUserId() (int, error) {
	user, err := GetCurrentUser(t.ctx)
	if err != nil {
		return 0, err
	}
	me := user.Me
	if len(me) == 0 {
		return 0, fmt.Errorf("Can't retrieve user information")
	}
	return me[0].Id, nil
}

func (t hardcoverTarget) collateIdentifiers(books []model.Book) (slugs []string, editions []int) {
	for _, book := range books {
		slug := book.Identifiers.Hardcover
		if slug != "" {
			slugs = append(slugs, slug)
		}
		edition := book.Identifiers.HardcoverEdition
		if edition != "" {
			editionId, err := strconv.Atoi(edition)
			if err != nil {
				log.Debug("Invalid edition id for book", "book", book.Name, "slug", slug, "edition", edition)
			} else {
				editions = append(editions, editionId)
			}
		}
	}
	return
}

func (t hardcoverTarget) ProcessReads(books []model.Book) ([]model.Book, error) {
	userId, err := t.getUserId()
	if err != nil {
		return nil, err
	}
	slugs, editions := t.collateIdentifiers(books)
	res, err := GetUserBooksBySlugOrEdition(t.ctx, slugs, editions, userId)
	if err != nil {
		return nil, err
	}
	log.Info("received response", "res", res)
	return books, nil
}

func (t hardcoverTarget) UpdateStatus(book model.Book) error {
	return nil
}

func New(ctx context.Context) internal.SyncTarget {
	return hardcoverTarget{
		ctx: ctx,
		Target: internal.Target{
			Name: "hardcover",
		},
	}
}
