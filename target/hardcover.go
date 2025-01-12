package target

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/config"
	"github.com/RobBrazier/readflow/source"
	"github.com/RobBrazier/readflow/target/hardcover"
	"github.com/charmbracelet/log"
)

//go:generate go run github.com/Khan/genqlient ../schemas/hardcover/genqlient.yaml

type HardcoverTarget struct {
	Target
	GraphQLTarget
	ctx    context.Context
	client graphql.Client
	log    *log.Logger
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
		t.client = t.GraphQLTarget.getClient(t.apiUrl, t.Token())
	}
	return t.client
}

func (t HardcoverTarget) Token() string {
	cfg := config.GetFromContext(t.ctx)
	return cfg.Tokens.Hardcover
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
	_, err := hardcover.FinishBookProgress(ctx, client, readId, pages, edition, startTime, finishTime)
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
	log := t.log.With("book", book.Current.BookName)

	log.Info("Retrieved book data", "localProgress", localProgress, "remoteProgress", remoteProgress)

	if localProgress <= remoteProgress {
		log.Info("Skipping update as target is already up-to-date")
		return nil
	}
	pages := float64(bookProgress.pages)
	progress := float64(localProgress / 100)
	newPagesCount := int(math.Round(pages * progress))

	if bookProgress.readId != nil {
		log.Info("Updating progress for", "pages", newPagesCount)
		startTime := time.Now()
		if bookProgress.startTime != nil {
			startTime = *bookProgress.startTime
		}
		if progress == 1 { // 100%
			log.Info("Marking book as finished", "book", book.Current.BookName)
			err := t.finishProgress(*bookProgress.readId, bookProgress.bookId, newPagesCount, bookProgress.edition, startTime)
			if err != nil {
				log.Error("error finishing book", "error", err)
			}
		} else {
			err := t.updateProgress(*bookProgress.readId, bookProgress.bookId, newPagesCount, bookProgress.edition, bookProgress.status, startTime)
			if err != nil {
				log.Error("error updating progress", "error", err)
			}
		}
	} else {
		log.Info("Starting progress for", "pages", newPagesCount)
		err := t.startProgress(bookProgress.bookId, newPagesCount, bookProgress.edition, bookProgress.status)
		if err != nil {
			t.log.Error("error starting progress", "error", err)
		}
	}

	return nil
}

func NewHardcoverTarget(ctx context.Context) SyncTarget {
	name := "hardcover"
	logger := log.WithPrefix(name)
	target := &HardcoverTarget{
		ctx: ctx,
		log: logger,
		Target: Target{
			name:   name,
			apiUrl: "https://api.hardcover.app/v1/graphql",
		},
	}
	return target
}
