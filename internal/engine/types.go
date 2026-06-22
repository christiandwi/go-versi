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
	ContentHash  string
	CreatedAt    time.Time
}

type ObjectSummary struct {
	ID           ObjectID
	RepositoryID RepositoryID
	Path         string
	ContentHash  string
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

type CommitChange struct {
	Path string
	Data []byte
}
