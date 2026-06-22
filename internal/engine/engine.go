package engine

import (
	"context"
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
