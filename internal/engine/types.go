package engine

import "time"

type RepositoryID string

type Repository struct {
	ID        RepositoryID
	Name      string
	CreatedAt time.Time
}
