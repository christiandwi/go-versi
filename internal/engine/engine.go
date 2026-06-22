package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Store interface {
	WithTx(ctx context.Context, fn func(Tx) error) error
}

type Tx interface {
	CreateRepository(ctx context.Context, repo Repository) error
	GetRepository(ctx context.Context, id RepositoryID) (Repository, error)
	CreateObject(ctx context.Context, obj Object) error
	GetObject(ctx context.Context, id ObjectID) (Object, error)
	GetObjectSummary(ctx context.Context, id ObjectID) (ObjectSummary, error)
	CreateCommit(ctx context.Context, commit Commit) error
	GetCommit(ctx context.Context, id CommitID) (Commit, error)
	SetRef(ctx context.Context, ref Ref) error
	GetRef(ctx context.Context, repoID RepositoryID, name string) (Ref, error)
}

type Engine struct {
	store Store
	now   func() time.Time
}

func New(store Store) *Engine {
	return &Engine{
		store: store,
		now:   time.Now,
	}
}

func (e *Engine) CreateRepository(ctx context.Context, name string) (Repository, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Repository{}, fmt.Errorf("%w: repository name is required", ErrValidation)
	}

	repo := Repository{
		ID:        HashRepository(name),
		Name:      name,
		CreatedAt: e.now().UTC(),
	}

	err := e.store.WithTx(ctx, func(tx Tx) error {
		return tx.CreateRepository(ctx, repo)
	})
	if err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (e *Engine) GetRepository(ctx context.Context, id RepositoryID) (Repository, error) {
	var repo Repository
	err := e.store.WithTx(ctx, func(tx Tx) error {
		var err error
		repo, err = tx.GetRepository(ctx, id)
		return err
	})
	return repo, err
}

func (e *Engine) CreateObject(ctx context.Context, repoID RepositoryID, objectPath string, data []byte) (Object, error) {
	if repoID == "" {
		return Object{}, fmt.Errorf("%w: repository id is required", ErrValidation)
	}

	objectPath = strings.TrimSpace(objectPath)
	if objectPath == "" {
		return Object{}, fmt.Errorf("%w: object path is required", ErrValidation)
	}

	objectID, err := NewObjectID()
	if err != nil {
		return Object{}, err
	}

	obj := Object{
		ID:           objectID,
		RepositoryID: repoID,
		Path:         objectPath,
		Data:         append([]byte(nil), data...),
		ContentHash:  HashObjectContent(data),
		CreatedAt:    e.now().UTC(),
	}

	err = e.store.WithTx(ctx, func(tx Tx) error {
		if _, err := tx.GetRepository(ctx, repoID); err != nil {
			return err
		}
		return tx.CreateObject(ctx, obj)
	})
	if err != nil {
		return Object{}, err
	}

	return obj, nil
}

func (e *Engine) GetObject(ctx context.Context, id ObjectID) (Object, error) {
	var obj Object
	err := e.store.WithTx(ctx, func(tx Tx) error {
		var err error
		obj, err = tx.GetObject(ctx, id)
		return err
	})
	return obj, err
}

func (e *Engine) CreateCommit(ctx context.Context, repoID RepositoryID, objectIDs []ObjectID, message string) (Commit, error) {
	if repoID == "" {
		return Commit{}, fmt.Errorf("%w: repository id is required", ErrValidation)
	}
	if len(objectIDs) == 0 {
		return Commit{}, fmt.Errorf("%w: at least one object is required", ErrValidation)
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return Commit{}, fmt.Errorf("%w: commit message is required", ErrValidation)
	}

	commitID, err := NewCommitID()
	if err != nil {
		return Commit{}, err
	}

	commit := Commit{
		ID:           commitID,
		RepositoryID: repoID,
		ObjectIDs:    append([]ObjectID(nil), objectIDs...),
		Message:      message,
		CreatedAt:    e.now().UTC(),
	}

	err = e.store.WithTx(ctx, func(tx Tx) error {
		if _, err := tx.GetRepository(ctx, repoID); err != nil {
			return err
		}

		for _, objectID := range objectIDs {
			obj, err := tx.GetObjectSummary(ctx, objectID)
			if err != nil {
				return err
			}
			if obj.RepositoryID != repoID {
				return fmt.Errorf("%w: object %q belongs to repository %q", ErrValidation, objectID, obj.RepositoryID)
			}
		}

		return tx.CreateCommit(ctx, commit)
	})
	if err != nil {
		return Commit{}, err
	}

	return commit, nil
}

func (e *Engine) GetCommit(ctx context.Context, id CommitID) (Commit, error) {
	var commit Commit
	err := e.store.WithTx(ctx, func(tx Tx) error {
		var err error
		commit, err = tx.GetCommit(ctx, id)
		return err
	})
	return commit, err
}

func (e *Engine) SetRef(ctx context.Context, repoID RepositoryID, name string, commitID CommitID) (Ref, error) {
	if repoID == "" {
		return Ref{}, fmt.Errorf("%w: repository id is required", ErrValidation)
	}
	if commitID == "" {
		return Ref{}, fmt.Errorf("%w: commit id is required", ErrValidation)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Ref{}, fmt.Errorf("%w: ref name is required", ErrValidation)
	}

	ref := Ref{
		RepositoryID: repoID,
		Name:         name,
		CommitID:     commitID,
		UpdatedAt:    e.now().UTC(),
	}

	err := e.store.WithTx(ctx, func(tx Tx) error {
		if _, err := tx.GetRepository(ctx, repoID); err != nil {
			return err
		}

		commit, err := tx.GetCommit(ctx, commitID)
		if err != nil {
			return err
		}
		if commit.RepositoryID != repoID {
			return fmt.Errorf("%w: commit %q belongs to repository %q", ErrValidation, commitID, commit.RepositoryID)
		}

		currentRef, err := tx.GetRef(ctx, repoID, name)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if err == nil {
			if currentRef.CommitID == commitID {
				return fmt.Errorf("%w: ref %q already points to commit %q", ErrNoChanges, name, commitID)
			}

			currentCommit, err := tx.GetCommit(ctx, currentRef.CommitID)
			if err != nil {
				return err
			}
			same, err := sameSnapshot(ctx, tx, currentCommit, commit)
			if err != nil {
				return err
			}
			if same {
				return fmt.Errorf("%w: ref %q already points to the same content", ErrNoChanges, name)
			}
		}

		return tx.SetRef(ctx, ref)
	})
	if err != nil {
		return Ref{}, err
	}

	return ref, nil
}

func (e *Engine) GetRef(ctx context.Context, repoID RepositoryID, name string) (Ref, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Ref{}, fmt.Errorf("%w: ref name is required", ErrValidation)
	}

	var ref Ref
	err := e.store.WithTx(ctx, func(tx Tx) error {
		var err error
		ref, err = tx.GetRef(ctx, repoID, name)
		return err
	})
	return ref, err
}

func (e *Engine) CommitToRef(ctx context.Context, repoID RepositoryID, refName string, changes []CommitChange, message string) (Commit, Ref, error) {
	if repoID == "" {
		return Commit{}, Ref{}, fmt.Errorf("%w: repository id is required", ErrValidation)
	}

	refName = strings.TrimSpace(refName)
	if refName == "" {
		return Commit{}, Ref{}, fmt.Errorf("%w: ref name is required", ErrValidation)
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return Commit{}, Ref{}, fmt.Errorf("%w: commit message is required", ErrValidation)
	}

	normalized, err := normalizeCommitChanges(changes)
	if err != nil {
		return Commit{}, Ref{}, err
	}

	now := e.now().UTC()
	var commit Commit
	var ref Ref

	err = e.store.WithTx(ctx, func(tx Tx) error {
		if _, err := tx.GetRepository(ctx, repoID); err != nil {
			return err
		}

		proposedSnapshot := make(map[string]string, len(normalized))
		for _, change := range normalized {
			proposedSnapshot[change.Path] = HashObjectContent(change.Data)
		}

		currentRef, err := tx.GetRef(ctx, repoID, refName)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if err == nil {
			currentCommit, err := tx.GetCommit(ctx, currentRef.CommitID)
			if err != nil {
				return err
			}
			currentSnapshot, err := snapshotObjects(ctx, tx, currentCommit.ObjectIDs)
			if err != nil {
				return err
			}
			if sameSnapshotMap(currentSnapshot, proposedSnapshot) {
				return fmt.Errorf("%w: ref %q already points to the same content", ErrNoChanges, refName)
			}
		}

		objectIDs := make([]ObjectID, 0, len(normalized))
		for _, change := range normalized {
			objectID, err := NewObjectID()
			if err != nil {
				return err
			}

			obj := Object{
				ID:           objectID,
				RepositoryID: repoID,
				Path:         change.Path,
				Data:         append([]byte(nil), change.Data...),
				ContentHash:  proposedSnapshot[change.Path],
				CreatedAt:    now,
			}
			if err := tx.CreateObject(ctx, obj); err != nil {
				return err
			}
			objectIDs = append(objectIDs, objectID)
		}

		commitID, err := NewCommitID()
		if err != nil {
			return err
		}

		commit = Commit{
			ID:           commitID,
			RepositoryID: repoID,
			ObjectIDs:    objectIDs,
			Message:      message,
			CreatedAt:    now,
		}
		if err := tx.CreateCommit(ctx, commit); err != nil {
			return err
		}

		ref = Ref{
			RepositoryID: repoID,
			Name:         refName,
			CommitID:     commitID,
			UpdatedAt:    now,
		}
		return tx.SetRef(ctx, ref)
	})
	if err != nil {
		return Commit{}, Ref{}, err
	}

	return commit, ref, nil
}

func sameSnapshot(ctx context.Context, tx Tx, a Commit, b Commit) (bool, error) {
	aObjects, err := snapshotObjects(ctx, tx, a.ObjectIDs)
	if err != nil {
		return false, err
	}
	bObjects, err := snapshotObjects(ctx, tx, b.ObjectIDs)
	if err != nil {
		return false, err
	}

	if len(aObjects) != len(bObjects) {
		return false, nil
	}

	return sameSnapshotMap(aObjects, bObjects), nil
}

func sameSnapshotMap(aObjects map[string]string, bObjects map[string]string) bool {
	if len(aObjects) != len(bObjects) {
		return false
	}

	for path, aHash := range aObjects {
		bHash, ok := bObjects[path]
		if !ok {
			return false
		}
		if aHash != bHash {
			return false
		}
	}

	return true
}

func snapshotObjects(ctx context.Context, tx Tx, objectIDs []ObjectID) (map[string]string, error) {
	objects := make(map[string]string, len(objectIDs))
	for _, objectID := range objectIDs {
		obj, err := tx.GetObjectSummary(ctx, objectID)
		if err != nil {
			return nil, err
		}
		objects[obj.Path] = obj.ContentHash
	}
	return objects, nil
}

func normalizeCommitChanges(changes []CommitChange) ([]CommitChange, error) {
	if len(changes) == 0 {
		return nil, fmt.Errorf("%w: at least one change is required", ErrValidation)
	}

	seen := make(map[string]struct{}, len(changes))
	normalized := make([]CommitChange, 0, len(changes))
	for _, change := range changes {
		path := strings.TrimSpace(change.Path)
		if path == "" {
			return nil, fmt.Errorf("%w: change path is required", ErrValidation)
		}
		if _, ok := seen[path]; ok {
			return nil, fmt.Errorf("%w: duplicate change path %q", ErrValidation, path)
		}
		seen[path] = struct{}{}
		normalized = append(normalized, CommitChange{
			Path: path,
			Data: append([]byte(nil), change.Data...),
		})
	}

	return normalized, nil
}
