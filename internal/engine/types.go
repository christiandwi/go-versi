package engine

import "time"

type RepositoryID string

type Repository struct {
	ID        RepositoryID
	Name      string
	CreatedAt time.Time
}

type ObjectID string

type Object struct {
	ID           ObjectID
	RepositoryID RepositoryID
	Path         string
	Data         []byte
	CreatedAt    time.Time
}

type CommitID string

type Commit struct {
	ID           CommitID
	RepositoryID RepositoryID
	ObjectIDs    []ObjectID
	Message      string
	CreatedAt    time.Time
}

type Ref struct {
	RepositoryID RepositoryID
	Name         string
	CommitID     CommitID
	UpdatedAt    time.Time
}
