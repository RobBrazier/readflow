package internal

import "github.com/RobBrazier/readflow/internal/model"

type SyncSource interface {
	GetName() string
	GetRecentReads() ([]model.Book, error)
}

type Source struct {
	Name string
	SyncSource
}

func (s Source) GetName() string {
	return s.Name
}
