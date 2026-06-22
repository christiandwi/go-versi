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
