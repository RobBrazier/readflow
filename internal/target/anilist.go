package target

import (
	"context"
	"math"
	"strconv"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target/anilist"
	"github.com/charmbracelet/log"
)

//go:generate go run github.com/Khan/genqlient ../../schemas/anilist/genqlient.yaml

type AnilistTarget struct {
	GraphQLTarget
	Target
	client graphql.Client
	ctx    context.Context
	log    *log.Logger
}

func dereferenceDefault[T any](pointer *T, defaultValue T) T {
	if pointer == nil {
		return defaultValue
	}
	return *pointer
}

func (t AnilistTarget) Login() (string, error) {
	return "https://anilist.co/api/v2/oauth/authorize?client_id=21288&response_type=token", nil
}

func (t *AnilistTarget) getClient() graphql.Client {
	if t.client == nil {
		t.client = t.GraphQLTarget.getClient(t.apiUrl, t.Token())
	}
	return t.client
}

func (t AnilistTarget) Token() string {
	cfg := config.GetFromContext(t.ctx)
	return cfg.Tokens.Anilist
}

func (t AnilistTarget) ShouldProcess(book source.BookContext) bool {
	id := book.Current.AnilistID
	if id == nil {
		return false
	}
	if *id == "" {
		return false
	}
	return true
}

func (t *AnilistTarget) getLocalVolumes(book source.Book, maxVolumes int) int {
	// lets just assume it's volume 1 if the pointer is null (i.e. no series)
	volume := dereferenceDefault(book.BookSeriesIndex, 1)

	if maxVolumes > 0 && volume > maxVolumes {
		t.log.Warn("Volume number exceeds the volume count on anilist - capping value", "book", book.BookName, "volume", volume, "max", maxVolumes)

	}

	return volume
}

func (t *AnilistTarget) getLocalChapters(book source.BookContext) (current int, previous int) {
	currentVolumeChapters := dereferenceDefault(book.Current.ChapterCount, 0)
	previousVolumeChapters := 0
	if len(book.Previous) > 0 {
		for _, book := range book.Previous {
			previousVolumeChapters += dereferenceDefault(book.ChapterCount, 0)
		}
	}
	return currentVolumeChapters, previousVolumeChapters
}

func (t *AnilistTarget) getEstimatedNewChapterCount(book source.BookContext, maxChapters int) int {
	chapter, localPreviousChapters := t.getLocalChapters(book)

	progress := dereferenceDefault(book.Current.ProgressPercent, 0.0) / 100
	latestVolumeChapter := int(math.Round(float64(chapter) * progress))

	estimatedChapter := localPreviousChapters + latestVolumeChapter

	t.log.Debug("Estimated current chapter", "book", book.Current.BookName, "progress", progress, "chapter", estimatedChapter)

	if maxChapters > 0 && estimatedChapter > maxChapters {
		t.log.Warn("Estimated chapter exceeds the chapter count on anilist - capping value", "book", book.Current.BookName, "estimated", estimatedChapter, "max", maxChapters)
		estimatedChapter = maxChapters
	}

	return estimatedChapter
}

func (t *AnilistTarget) UpdateReadStatus(book source.BookContext) error {
	anilistId, err := strconv.Atoi(*book.Current.AnilistID)
	if err != nil {
		t.log.Error("Invalid anilist id", "id", *book.Current.AnilistID)
		return err
	}
	ctx := t.ctx
	client := t.getClient()
	current, err := anilist.GetUserMediaById(ctx, client, anilistId)
	if err != nil {
		return err
	}

	bookName := book.Current.BookName
	title := current.Media.Title.UserPreferred
	maxVolumes := current.Media.Volumes
	maxChapters := current.Media.Chapters
	status := current.Media.MediaListEntry.Status

	remoteVolumes := current.Media.MediaListEntry.ProgressVolumes
	remoteChapters := current.Media.MediaListEntry.Progress

	localVolumes := t.getLocalVolumes(book.Current, maxVolumes)
	estimatedChapter := t.getEstimatedNewChapterCount(book, maxChapters)

	if localVolumes <= remoteVolumes && estimatedChapter <= remoteChapters {
		t.log.
			With("book", bookName, "title", title).
			Info("Skipping update as target is already up-to-date")
		return nil
	}
	if status == "" {
		status = anilist.MediaListStatusCurrent
	}
	if estimatedChapter == maxChapters {
		status = anilist.MediaListStatusCompleted
	}
	t.log.Info("Updating progress for", "book", bookName, "volume", localVolumes, "chapter", estimatedChapter)
	_, err = anilist.UpdateProgress(ctx, client, anilistId, estimatedChapter, localVolumes, status)
	if err != nil {
		t.log.Error("error updating progress", "error", err)
		return err
	}

	return nil
}

func NewAnilistTarget(ctx context.Context) SyncTarget {
	name := "anilist"
	logger := log.WithPrefix(name)
	target := &AnilistTarget{
		ctx: ctx,
		log: logger,
		Target: Target{
			name:   name,
			apiUrl: "https://graphql.anilist.co",
		},
	}
	return target
}
