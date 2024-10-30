package source

type Source interface {
	GetChaptersColumn() (string, error)
}
