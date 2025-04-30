package hardcover

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/model"
	"github.com/charmbracelet/log"
)

//go:generate go run github.com/Khan/genqlient ../../schemas/hardcover/genqlient.yaml

type hardcoverTarget struct {
	ctx context.Context
	internal.Target
}

type ReadStatus int

const (
	STATUS_WANT_TO_READ   ReadStatus = 1
	STATUS_IN_PROGRESS    ReadStatus = 2
	STATUS_FINISHED       ReadStatus = 3
	STATUS_PAUSED         ReadStatus = 4
	STATUS_DID_NOT_FINISH ReadStatus = 5
)

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

func (t hardcoverTarget) collateIdentifiers(books []model.Book) ([]string, []int) {
	slugs := []string{}
	editions := []int{}
	for _, book := range books {
		slug := book.Identifiers.Hardcover
		if slug != "" {
			slugs = append(slugs, slug)
		}
		edition := book.Identifiers.HardcoverEdition
		if edition != "" {
			editionId, err := strconv.Atoi(edition)
			if err != nil {
				log.Warn("Invalid edition id for book", "book", book.Name, "slug", slug, "edition", edition)
			} else {
				editions = append(editions, editionId)
			}
		}
	}
	return slugs, editions
}

func (t hardcoverTarget) matchRemoteBook(slug string, books []GetUserBooksBySlugOrEditionBooks) *GetUserBooksBySlugOrEditionBooks {
	for _, book := range books {
		if book.Slug == slug {
			return &book
		}
	}
	return nil
}

func (t hardcoverTarget) updateFromRemote(book *model.Book, hardcoverBook *GetUserBooksBySlugOrEditionBooks) {
	books := hardcoverBook.User_books
	editions := hardcoverBook.Editions

	book.Metadata = make(map[string]any)
	book.Metadata["bookId"] = hardcoverBook.Id
	book.Metadata["status"] = STATUS_WANT_TO_READ
	book.Metadata["startedAt"] = time.Now()
	if len(editions) > 0 {
		// grab the number of pages from the linked edition
		// this edition will either be the one matching the book's edition id or the most popular one
		book.Metadata["pages"] = editions[0].Pages
	}

	if len(books) == 0 {
		// the book isn't in any existing list, can't fetch any more data
		return
	}
	userBook := books[0]
	status := ReadStatus(userBook.Status_id)
	book.Metadata["userBookId"] = userBook.Id
	reads := userBook.User_book_reads
	if len(reads) == 0 {
		// it's in a TBR list, but not started reading yet
		return
	}
	latestRead := reads[0]
	book.Progress.Remote = float64(latestRead.Progress)
	book.Identifiers.HardcoverEdition = strconv.Itoa(latestRead.Edition.Id)
	book.Metadata["readId"] = latestRead.Id
	// if we don't have the edition id
	if len(editions) == 0 {
		book.Metadata["pages"] = latestRead.Edition.Pages
	}
	book.Metadata["status"] = status
	book.Metadata["startedAt"] = latestRead.Started_at
}

func (t hardcoverTarget) enhanceProcessableBooks(books []model.Book, hardcoverResponse GetUserBooksBySlugOrEditionResponse) []model.Book {
	processable := []model.Book{}
	for _, book := range books {
		slug := book.Identifiers.Hardcover
		edition := book.Identifiers.HardcoverEdition
		hardcoverBook := t.matchRemoteBook(slug, hardcoverResponse.Books)
		if hardcoverBook == nil {
			continue
		}
		// does the user already have it in one of their lists?
		userBooks := hardcoverBook.User_books
		// do we have the edition information already? if not skip
		if len(userBooks) == 0 && edition == "" {
			continue
		}

		t.updateFromRemote(&book, hardcoverBook)

		localProgress := math.Round(book.Progress.Local)
		remoteProgress := math.Round(book.Progress.Remote)
		status := book.Metadata["status"].(ReadStatus)
		if localProgress == remoteProgress {
			if localProgress == 100 && status != STATUS_FINISHED {
				log.Info("Book is complete, but missing finished status", "book", slug)
			} else if status != STATUS_IN_PROGRESS {
				log.Info("Book is in progress, but missing in progress status", "book", slug)
			} else {
				log.Info("Book is already up-to-date", "book", slug, "progress", book.Progress.Local)
				continue
			}
		}
		processable = append(processable, book)
	}
	return processable
}

func (t hardcoverTarget) ProcessReads(books []model.Book) ([]model.Book, error) {
	slugs, editions := t.collateIdentifiers(books)
	if len(slugs) == 0 && len(editions) == 0 {
		log.Debug("no eligible books - skipping")
		return nil, nil
	}
	userId, err := t.getUserId()
	if err != nil {
		return nil, err
	}
	log.Debug("will query hardcover for books", "slugs", slugs, "editions", editions)
	res, err := GetUserBooksBySlugOrEdition(t.ctx, slugs, editions, userId)
	if err != nil {
		return nil, err
	}
	books = t.enhanceProcessableBooks(books, *res)
	return books, nil
}

func (t hardcoverTarget) populateReadId(book model.Book) model.Book {
	if _, ok := book.Metadata["readId"]; ok {
		return book
	}
	bookId := book.Metadata["bookId"].(int)
	editionId, _ := strconv.Atoi(book.Identifiers.HardcoverEdition)
	res, err := CreateUserBook(t.ctx, bookId, int(STATUS_IN_PROGRESS), editionId)
	if err != nil {
		log.Error("couldn't add book to want to read", "error", err)
		return book
	}
	userBookRead := res.Insert_user_book.User_book.User_book_reads[0]
	book.Metadata["startedAt"] = userBookRead.Started_at
	book.Metadata["userBookId"] = res.Insert_user_book.Id
	book.Metadata["readId"] = userBookRead.Id
	book.Metadata["status"] = STATUS_IN_PROGRESS
	log.Debug("creating new readId and populated", "book", book)
	return book
}

func (t hardcoverTarget) ensureBookStatus(bookStatus ReadStatus, userBookId int, expectedStatus ReadStatus) ReadStatus {
	if bookStatus != expectedStatus {
		log.Debug("book status doesn't match expected value", "expected", expectedStatus, "actual", bookStatus)
		log.Debug("Changing book status", "userBookId", userBookId, "status", expectedStatus)
		res, err := ChangeBookStatus(t.ctx, userBookId, int(expectedStatus))
		log.Debug("response", "res", res)
		if err != nil {
			log.Error("couldn't change status to", "status", expectedStatus)
			return bookStatus
		}
	}
	return expectedStatus
}

func (t hardcoverTarget) UpdateStatus(book model.Book) error {
	localProgress := math.Round(book.Progress.Local)
	remoteProgress := math.Round(book.Progress.Remote)

	totalPages := book.Metadata["pages"].(int)
	// invalid editions have already been filtered out by this point
	edition, _ := strconv.Atoi(book.Identifiers.HardcoverEdition)
	book = t.populateReadId(book)
	status := book.Metadata["status"].(ReadStatus)
	userBookId := book.Metadata["userBookId"].(int)
	readId := book.Metadata["readId"].(int)
	startedAt := book.Metadata["startedAt"].(time.Time)

	// handle finished books
	if localProgress == remoteProgress && localProgress == 100 {
		if status != STATUS_FINISHED {
			// change status
		}
		// update progress (probably not needed for finished)
		return nil
	}

	status = t.ensureBookStatus(status, userBookId, STATUS_IN_PROGRESS)
	// handle in progress read updates
	if localProgress > remoteProgress {
		pages := int(math.Round(float64(totalPages) * (localProgress / 100)))
		if status != STATUS_IN_PROGRESS {
			// change status
		}
		// update an existing read
		res, err := UpdateBookProgress(t.ctx, readId, pages, edition, startedAt)
		log.Debug("updated progress", "book", book.Name, "readId", readId, "pages", pages, "edition", edition, "res", res)
		if err != nil {
			log.Error("couldn't update progress", "error", err)
		}
	}
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
