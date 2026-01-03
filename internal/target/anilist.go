package target

import (
	"context"
	"math"
	"strconv"

	"github.com/Khan/genqlient/graphql"
	"github.com/RobBrazier/readflow/internal/config"
	"github.com/RobBrazier/readflow/internal/source"
	"github.com/RobBrazier/readflow/internal/target/anilist"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:generate go tool genqlient ../../schemas/anilist/genqlient.yaml

type AnilistTarget struct {
	GraphQLTarget
	Target
	client graphql.Client
	ctx    context.Context
	log    zerolog.Logger
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
		t.client = t.GraphQLTarget.getClient(t.ApiUrl, t.GetToken())
	}
	return t.client
}

func (t AnilistTarget) GetToken() string {
	return config.Get().String("target.hardcover.token")
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

func (t *AnilistTarget) GetCurrentUser() string {
	response, err := anilist.GetCurrentUser(t.ctx, t.getClient())
	cobra.CheckErr(err)
	return response.Viewer.Name
}

func (t *AnilistTarget) getLocalVolumes(book source.Book, maxVolumes int) int {
	// lets just assume it's volume 1 if the pointer is null (i.e. no series)
	volume := dereferenceDefault(book.BookSeriesIndex, 1)

	if maxVolumes > 0 && volume > maxVolumes {
		t.log.Warn().
			Str("book", book.BookName).
			Int("volume", volume).
			Int("max", maxVolumes).
			Msg("Volume number exceeds the volume count on anilist - capping value")
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

	log := t.log.With().Str("book", book.Current.BookName).Logger()

	log.Debug().
		Float64("progress", progress).
		Int("chapter", estimatedChapter).
		Msg("Estimated current chapter")

	if maxChapters > 0 && estimatedChapter > maxChapters {
		log.Warn().
			Int("estimated", estimatedChapter).
			Int("max", maxChapters).
			Msg("Estimated chapter exceeds the chapter count on anilist - capping value")
		estimatedChapter = maxChapters
	}

	return estimatedChapter
}

func (t *AnilistTarget) UpdateReadStatus(book source.BookContext) error {
	anilistId, err := strconv.Atoi(*book.Current.AnilistID)
	if err != nil {
		t.log.Error().Str("id", *book.Current.AnilistID).Msg("Invalid anilist id")
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

	log := t.log.With().Str("book", bookName).Logger()

	if localVolumes <= remoteVolumes && estimatedChapter <= remoteChapters {
		log.Info().Str("title", title).Msg("Skipping update as target is already up-to-date")
		return nil
	}
	if status == "" {
		status = anilist.MediaListStatusCurrent
	}
	if estimatedChapter == maxChapters {
		status = anilist.MediaListStatusCompleted
	}
	log.Info().
		Int("volume", localVolumes).
		Int("chapter", estimatedChapter).
		Msg("Updating progress for")

	_, err = anilist.UpdateProgress(ctx, client, anilistId, estimatedChapter, localVolumes, status)
	if err != nil {
		log.Error().Err(err).Msg("error updating progress")
		return err
	}

	return nil
}

func NewAnilistTarget(ctx context.Context) SyncTarget {
	name := "anilist"
	logger := log.With().Str("target", name).Logger()
	target := &AnilistTarget{
		ctx: ctx,
		log: logger,
		Target: Target{
			Name:   name,
			ApiUrl: "https://graphql.anilist.co",
		},
	}
	return target
}
