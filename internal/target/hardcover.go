package target

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target/hardcover"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:generate go run github.com/Khan/genqlient ../../schemas/hardcover/genqlient.yaml

type HardcoverTarget struct {
	Target
	GraphQLTarget
	ctx    context.Context
	client graphql.Client
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

func (t *HardcoverTarget) updateProgress(readId, bookId, pages, edition, status int, startTime time.Time) error {
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

func (t *HardcoverTarget) finishProgress(readId, bookId, pages, edition int, startTime time.Time) error {
	finishTime := time.Now()
	ctx := t.ctx
	client := t.getClient()
	_, err := hardcover.FinishBookProgress(ctx, client, bookId, pages, edition, startTime, finishTime)
	if err != nil {
		return err
	}
	_, err = hardcover.ChangeBookStatus(ctx, client, readId, 3) // finished
	return err
}

func (t *HardcoverTarget) startProgress(id, pages, edition, status int) error {
	startTime := time.Now()
	ctx := t.ctx
	client := t.getClient()
	if status != 2 { // in progress
		_, err := hardcover.ChangeBookStatus(ctx, client, id, 2)
		if err != nil {
			return err
		}
	}
	_, err := hardcover.StartBookProgress(ctx, client, id, pages, edition, startTime)
	return err
}

func (t *HardcoverTarget) UpdateReadStatus(book source.BookContext) error {
	slug := book.Current.HardcoverID
	if slug == nil {
		return BookNotFound
	}
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
	// round to 2 decimal places
	remoteProgress := math.Round(float64(bookProgress.progress)*100) / 100

	slog.Info("Got book data", "book", book.Current.BookName, "localProgress", localProgress, "remoteProgress", remoteProgress)

	if localProgress <= remoteProgress {
		slog.Info("Progress already up-to-date, skipping")
		return nil
	}
	pages := float64(bookProgress.pages)
	progress := float64(localProgress / 100)
	newPagesCount := int(math.Round(pages * progress))

	if bookProgress.readId != nil {
		slog.Info("Updating progress for", "book", book.Current.BookName, "pages", newPagesCount)
		startTime := time.Now()
		if bookProgress.startTime != nil {
			startTime = *bookProgress.startTime
		}
		if progress == 100.0 {
			err := t.finishProgress(*bookProgress.readId, bookProgress.bookId, newPagesCount, bookProgress.edition, startTime)
			if err != nil {
				slog.Error("error finishing book", "error", err)
			}
		} else {
			err := t.updateProgress(*bookProgress.readId, bookProgress.bookId, newPagesCount, bookProgress.edition, bookProgress.status, startTime)
			if err != nil {
				slog.Error("error updating progress", "error", err)
			}
		}
	} else {
		slog.Info("Starting progress for", "book", book.Current.BookName, "pages", newPagesCount)
		err := t.startProgress(bookProgress.bookId, newPagesCount, bookProgress.edition, bookProgress.status)
		if err != nil {
			slog.Error("error starting progress", "error", err)
		}
	}

	return nil
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
