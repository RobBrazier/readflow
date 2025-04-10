package internal

import "github.com/RobBrazier/readflow/internal/model"

type SyncTarget interface {
	GetName() string
	Login() (string, error)
	ProcessReads(books []model.Book) ([]model.Book, error)
	UpdateStatus(book model.Book) error
}

type Target struct {
	Name string
	SyncTarget
}

func (t Target) GetName() string {
	return t.Name
}
