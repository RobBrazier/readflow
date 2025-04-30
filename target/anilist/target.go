package anilist

import (
	"context"
	"slices"
	"strconv"

	"github.com/RobBrazier/readflow/internal"
	"github.com/RobBrazier/readflow/internal/model"
	"github.com/charmbracelet/log"
)

//go:generate go run github.com/Khan/genqlient ../../schemas/anilist/genqlient.yaml

type anilistTarget struct {
	internal.Target
	ctx context.Context
}

type ReadStatus MediaListStatus

const (
	STATUS_WANT_TO_READ ReadStatus = ReadStatus(MediaListStatusPlanning)
	STATUS_IN_PROGRESS  ReadStatus = ReadStatus(MediaListStatusCurrent)
	STATUS_REPEATING    ReadStatus = ReadStatus(MediaListStatusRepeating)
	STATUS_FINISHED     ReadStatus = ReadStatus(MediaListStatusCompleted)
	STATUS_PAUSED       ReadStatus = ReadStatus(MediaListStatusPaused)
	STATUS_DROPPED      ReadStatus = ReadStatus(MediaListStatusDropped)
)

func (t anilistTarget) Login() (string, error) {
	return "https://anilist.co/api/v2/oauth/authorize?client_id=21288&response_type=token", nil
}

func (t anilistTarget) collateIdentifiers(books []model.Book) []int {
	identifiers := []int{}
	for _, book := range books {
		identifier := book.Identifiers.Anilist
		if identifier != "" {
			mediaId, err := strconv.Atoi(identifier)
			if err != nil {
				log.Warn("Invalid media id for book", "book", book.Name, "identifier", identifier)
			} else {
				if slices.Contains(identifiers, mediaId) {
					continue
				}
				identifiers = append(identifiers, mediaId)
			}
		}
	}
	return identifiers
}

func (t anilistTarget) matchRemoteBook(identifier string, books []GetUserMediaByIdsPageMedia) *GetUserMediaByIdsPageMedia {
	for _, book := range books {
		if strconv.Itoa(book.Id) == identifier {
			return &book
		}
	}
	return nil
}

func (t anilistTarget) filterPreviousBooks(books []model.Book) []model.Book {
	filtered := []model.Book{}
	seriesBooks := map[int][]model.Book{}
	for _, book := range books {
		seriesBooks[book.Series.ID] = append(seriesBooks[book.Series.ID], book)
	}

	for _, books := range seriesBooks {
		maxIndex := 0
		var latestBook model.Book
		for _, book := range books {
			if maxIndex < book.Series.Index {
				maxIndex = book.Series.Index
				latestBook = book
			}
		}
		filtered = append(filtered, latestBook)
	}
	return filtered
}

func (t anilistTarget) enhanceProcessableBooks(books []model.Book, anilistResponse GetUserMediaByIdsResponse) []model.Book {
	processable := []model.Book{}
	// remove any previous books in a series to avoid unnecessary updates
	books = t.filterPreviousBooks(books)
	for _, book := range books {
		identifier := book.Identifiers.Anilist
		anilistEntry := t.matchRemoteBook(identifier, anilistResponse.Page.Media)
		if anilistEntry == nil {
			continue
		}
		t.updateFromRemote(&book, anilistEntry.MediaEntry)
		currentVolume := book.Series.Index
		currentChapter := book.CalculateTotalChapters()

		remoteVolume := book.Metadata["remoteVolume"].(int)
		remoteChapter := book.Metadata["remoteChapter"].(int)

		status := book.Metadata["status"].(ReadStatus)
		if currentVolume < remoteVolume {
			log.Debug("Book volume is behind remote - skipping", "book", book.Name, "currentVolume", currentVolume, "remoteVolume", remoteVolume)
			continue
		}
		if currentChapter < remoteChapter {
			log.Debug("Book chapter is behind remote - skipping", "book", book.Name, "currentChapter", currentChapter, "remoteChapter", remoteChapter)
			continue
		}

		if currentChapter == remoteChapter {
			// mark as complete if reached chapter count
			if status == STATUS_WANT_TO_READ {
				log.Info("Book is in progress, but missing in progress status", "mediaEntry", identifier)
			} else {
				log.Info("Book is already up-to-date", "book", identifier)
				continue
			}
		}

		processable = append(processable, book)
	}
	return processable
}

func (t anilistTarget) updateFromRemote(book *model.Book, entry MediaEntry) {
	mediaListEntry := entry.MediaListEntry
	totalVolumes := entry.Volumes
	totalChapters := entry.Chapters

	book.Metadata = make(map[string]any)
	book.Metadata["totalVolumes"] = totalVolumes
	book.Metadata["totalChapters"] = totalChapters

	if mediaListEntry.Status == "" {
		book.Metadata["status"] = STATUS_WANT_TO_READ
		return
	}
	book.Metadata["status"] = ReadStatus(mediaListEntry.Status)
	progressVolumes := mediaListEntry.ProgressVolumes
	progressChapters := mediaListEntry.Progress
	book.Metadata["remoteVolume"] = progressVolumes
	book.Metadata["remoteChapter"] = progressChapters
}

func (t anilistTarget) ProcessReads(books []model.Book) ([]model.Book, error) {
	identifiers := t.collateIdentifiers(books)
	if len(identifiers) == 0 {
		log.Debug("no eligible books - skipping")
		return nil, nil
	}
	log.Debug("will query anilist for books", "identifiers", identifiers)
	res, err := GetUserMediaByIds(t.ctx, identifiers)
	if err != nil {
		return nil, err
	}
	books = t.enhanceProcessableBooks(books, *res)
	return books, nil
}

func (t anilistTarget) UpdateStatus(book model.Book) error {
	mediaId, _ := strconv.Atoi(book.Identifiers.Anilist)

	currentVolume := book.Series.Index
	currentChapter := book.CalculateTotalChapters()

	status := book.Metadata["status"].(ReadStatus)
	totalChapters := book.Metadata["totalChapters"].(int)

	if currentChapter == totalChapters && totalChapters != 0 && status != STATUS_FINISHED {
		// set status to finished if it's completed but not marked like that
		status = STATUS_FINISHED
	}
	// set status to IN_PROGRESS if chapters have increased

	res, err := UpdateProgress(t.ctx, mediaId, currentChapter, currentVolume, MediaListStatus(status))
	log.Debug("updated progress", "book", book.Name, "volume", currentVolume, "chapter", currentChapter, "response", res, "err", err)
	if err != nil {
		log.Error("couldn't update progress", "error", err)
	}
	return nil
}

func New(ctx context.Context) internal.SyncTarget {
	return anilistTarget{
		ctx: ctx,
		Target: internal.Target{
			Name: "anilist",
		},
	}
}
