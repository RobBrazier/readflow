package target

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target/hardcover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:generate go tool genqlient ../../schemas/hardcover/genqlient.yaml

type HardcoverTarget struct {
	Target
	GraphQLTarget
	config config.HardcoverConfig
	ctx    context.Context
	client graphql.Client
	log    zerolog.Logger
}

type hardcoverProgress struct {
	bookId    int
	readId    *int
	status    int
	pages     int
	progress  float32
	startTime *time.Time
	edition   int
}

func (t HardcoverTarget) Login() (string, error) {
	return "https://hardcover.app/account/api", nil
}

func (t *HardcoverTarget) getClient() graphql.Client {
	if t.client == nil {
		t.client = t.GraphQLTarget.getClient(t.ApiUrl, t.GetToken())
	}
	return t.client
}

func (t HardcoverTarget) GetToken() string {
	return t.config.Token
}

func (t *HardcoverTarget) GetCurrentUser() string {
	response, err := hardcover.GetCurrentUser(t.ctx, t.getClient())
	cobra.CheckErr(err)
	return response.GetMe()[0].GetUsername()
}

func (t HardcoverTarget) ShouldProcess(book source.BookContext) bool {
	id := book.Current.HardcoverID
	if id == nil {
		return false
	}
	if *id == "" {
		return false
	}
	return true
}

// Yes this is absolutely horrible, but the generated code is horrible too...
func (t *HardcoverTarget) getCurrentBookProgress(slug string) (*hardcoverProgress, error) {
	current, err := hardcover.GetUserBooksBySlug(t.ctx, t.getClient(), slug)
	if err != nil {
		return nil, err
	}
	me := current.Me[0]
	userBooks := me.User_books

	if len(userBooks) == 0 {
		return nil, errors.New("Book not found in User Books - Skipping")
	}
	userBook := userBooks[0]
	status := userBook.Status_id
	reads := userBook.User_book_reads
	pages := userBook.Edition.Pages
	bookId := userBook.Book_id
	result := hardcoverProgress{
		bookId:  bookId,
		status:  status,
		pages:   pages,
		edition: userBook.Edition.Id,
	}
	if len(reads) == 0 {
		// book hasn't been started yet - assuming 0 progress
		return &result, nil
	}
	read := reads[0]
	id := read.Id
	result.readId = &id
	if read.Edition.Id != 0 {
		result.edition = read.Edition.Id
		result.pages = read.Edition.Pages
	}

	result.startTime = &read.Started_at
	progress := read.Progress
	result.progress = progress
	return &result, nil
}

func (t *HardcoverTarget) updateProgress(
	readId, bookId, pages, edition, status int,
	startTime time.Time,
) error {
	ctx := t.ctx
	client := t.getClient()
	if status != 2 { // in progress
		_, err := hardcover.ChangeBookStatus(ctx, client, bookId, 2)
		if err != nil {
			return err
		}
	}
	_, err := hardcover.UpdateBookProgress(ctx, client, readId, pages, edition, startTime)
	return err
}

func (t *HardcoverTarget) finishProgress(
	readId, bookId, pages, edition int,
	startTime time.Time,
) error {
	finishTime := time.Now()
	ctx := t.ctx
	client := t.getClient()
	_, err := hardcover.FinishBookProgress(
		ctx,
		client,
		readId,
		pages,
		edition,
		startTime,
		finishTime,
	)
	if err != nil {
		return err
	}
	_, err = hardcover.ChangeBookStatus(ctx, client, bookId, 3) // finished
	return err
}

func (t *HardcoverTarget) startProgress(bookId, pages, edition, status int) error {
	startTime := time.Now()
	ctx := t.ctx
	client := t.getClient()
	if status != 2 { // in progress
		_, err := hardcover.ChangeBookStatus(ctx, client, bookId, 2)
		if err != nil {
			return err
		}
	}
	_, err := hardcover.StartBookProgress(ctx, client, bookId, pages, edition, startTime)
	return err
}

func (t *HardcoverTarget) UpdateReadStatus(book source.BookContext) error {
	slug := book.Current.HardcoverID
	localProgressPointer := book.Current.ProgressPercent
	if localProgressPointer == nil {
		// no error, but nothing to update as we have no progress
		return nil
	}
	localProgress := *localProgressPointer
	bookProgress, err := t.getCurrentBookProgress(*slug)
	if err != nil {
		return err
	}
	// round to 0 decimal places to match the source progress
	remoteProgress := math.Round(float64(bookProgress.progress))
	log := t.log.With().Str("book", book.Current.BookName).Logger()

	log.Info().
		Float64("localProgress", localProgress).
		Float64("remoteProgress", remoteProgress).
		Msg("Retrieved book data")

	if localProgress <= remoteProgress {
		log.Info().Msg("Skipping update as target is already up-to-date")
		return nil
	}
	pages := float64(bookProgress.pages)
	progress := float64(localProgress / 100)
	newPagesCount := int(math.Round(pages * progress))

	if bookProgress.readId != nil {
		log.Info().Int("pages", newPagesCount).Msg("Updating progress for")
		startTime := time.Now()
		if bookProgress.startTime != nil {
			startTime = *bookProgress.startTime
		}
		if progress == 1 { // 100%
			log.Info().Str("book", book.Current.BookName).Msg("Marking book as finished")
			err := t.finishProgress(
				*bookProgress.readId,
				bookProgress.bookId,
				newPagesCount,
				bookProgress.edition,
				startTime,
			)
			if err != nil {
				log.Error().Err(err).Msg("error finishing book")
			}
		} else {
			err := t.updateProgress(*bookProgress.readId, bookProgress.bookId, newPagesCount, bookProgress.edition, bookProgress.status, startTime)
			if err != nil {
				log.Error().Err(err).Msg("error updating progress")
			}
		}
	} else {
		log.Info().Int("pages", newPagesCount).Msg("Starting progress for")
		err := t.startProgress(bookProgress.bookId, newPagesCount, bookProgress.edition, bookProgress.status)
		if err != nil {
			t.log.Error().Err(err).Msg("error starting progress")
		}
	}

	return nil
}

func NewHardcoverTarget(ctx context.Context, config config.HardcoverConfig) SyncTarget {
	name := "hardcover"
	logger := log.With().Str("target", name).Logger()
	target := &HardcoverTarget{
		config: config,
		ctx:    ctx,
		log:    logger,
		Target: Target{
			Name:   name,
			ApiUrl: "https://api.hardcover.app/v1/graphql",
		},
	}
	return target
}
