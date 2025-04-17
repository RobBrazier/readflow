package model

import "math"

type BookIdentifiers struct {
	Anilist          string
	Hardcover        string
	HardcoverEdition string
}

type BookProgress struct {
	Local  float64
	Remote float64
}

type BookSeries struct {
	ID      int
	Index   int
	Entries []Book
}

type Book struct {
	ID           int
	Name         string
	Identifiers  BookIdentifiers
	Progress     BookProgress
	ChapterCount int
	Series       BookSeries
	Metadata     map[string]any
}

func (b Book) getCurrentChapters() int {
	// progress percentage comes in like 50.0, converting to 0.5
	progress := b.Progress.Local / 100
	pages := b.ChapterCount
	if pages == 0 {
		// no chapters registered, can't calculate progress
		return 0
	}
	return int(math.Round(float64(pages) * progress))
}

func (b Book) CalculateTotalChapters() int {
	previousBookChapters := 0
	for _, book := range b.Series.Entries {
		previousBookChapters += book.ChapterCount
	}
	currentBookChapters := b.getCurrentChapters()
	return previousBookChapters + currentBookChapters
}

func (b *Book) AddBooksToSeries(books []Book) {
	b.Series.Entries = append(b.Series.Entries, books...)
}

func (b *Book) SetRemoteProgress(progress float64) {
	b.Progress.Remote = progress
}
