package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go-versi/internal/engine"
)

var _ engine.Store = (*Store)(nil)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) WithTx(ctx context.Context, fn func(engine.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	wrapped := &storeTx{tx: tx}
	if err := fn(wrapped); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

type storeTx struct {
	tx *sql.Tx
}

func (tx *storeTx) CreateRepository(ctx context.Context, repo engine.Repository) error {
	result, err := tx.tx.ExecContext(ctx, `
		INSERT INTO repositories (id, name, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, repo.ID, repo.Name, repo.CreatedAt)
	if err != nil {
		return fmt.Errorf("create repository: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("create repository rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: repository %q already exists", engine.ErrConflict, repo.ID)
	}
	return nil
}

func (tx *storeTx) GetRepository(ctx context.Context, id engine.RepositoryID) (engine.Repository, error) {
	var repo engine.Repository
	err := tx.tx.QueryRowContext(ctx, `
		SELECT id, name, created_at
		FROM repositories
		WHERE id = $1
	`, id).Scan(&repo.ID, &repo.Name, &repo.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return engine.Repository{}, fmt.Errorf("%w: repository %q", engine.ErrNotFound, id)
	}
	if err != nil {
		return engine.Repository{}, fmt.Errorf("get repository: %w", err)
	}
	return repo, nil
}
