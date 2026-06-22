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

func (tx *storeTx) CreateObject(ctx context.Context, obj engine.Object) error {
	result, err := tx.tx.ExecContext(ctx, `
		INSERT INTO objects (id, repository_id, path, data, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, obj.ID, obj.RepositoryID, obj.Path, obj.Data, obj.CreatedAt)
	if err != nil {
		return fmt.Errorf("create object: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("create object rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: object %q already exists", engine.ErrConflict, obj.ID)
	}
	return nil
}

func (tx *storeTx) GetObject(ctx context.Context, id engine.ObjectID) (engine.Object, error) {
	var obj engine.Object
	err := tx.tx.QueryRowContext(ctx, `
		SELECT id, repository_id, path, data, created_at
		FROM objects
		WHERE id = $1
	`, id).Scan(&obj.ID, &obj.RepositoryID, &obj.Path, &obj.Data, &obj.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return engine.Object{}, fmt.Errorf("%w: object %q", engine.ErrNotFound, id)
	}
	if err != nil {
		return engine.Object{}, fmt.Errorf("get object: %w", err)
	}
	return obj, nil
}

func (tx *storeTx) CreateCommit(ctx context.Context, commit engine.Commit) error {
	result, err := tx.tx.ExecContext(ctx, `
		INSERT INTO commits (id, repository_id, message, author, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, commit.ID, commit.RepositoryID, commit.Message, "", commit.CreatedAt)
	if err != nil {
		return fmt.Errorf("create commit: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("create commit rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: commit %q already exists", engine.ErrConflict, commit.ID)
	}

	for i, objectID := range commit.ObjectIDs {
		_, err := tx.tx.ExecContext(ctx, `
			INSERT INTO commit_objects (commit_id, object_id, position)
			VALUES ($1, $2, $3)
		`, commit.ID, objectID, i)
		if err != nil {
			return fmt.Errorf("create commit object: %w", err)
		}
	}

	return nil
}

func (tx *storeTx) GetCommit(ctx context.Context, id engine.CommitID) (engine.Commit, error) {
	var commit engine.Commit
	err := tx.tx.QueryRowContext(ctx, `
		SELECT id, repository_id, message, created_at
		FROM commits
		WHERE id = $1
	`, id).Scan(&commit.ID, &commit.RepositoryID, &commit.Message, &commit.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return engine.Commit{}, fmt.Errorf("%w: commit %q", engine.ErrNotFound, id)
	}
	if err != nil {
		return engine.Commit{}, fmt.Errorf("get commit: %w", err)
	}

	rows, err := tx.tx.QueryContext(ctx, `
		SELECT object_id
		FROM commit_objects
		WHERE commit_id = $1
		ORDER BY position ASC
	`, id)
	if err != nil {
		return engine.Commit{}, fmt.Errorf("get commit objects: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var objectID engine.ObjectID
		if err := rows.Scan(&objectID); err != nil {
			return engine.Commit{}, fmt.Errorf("scan commit object: %w", err)
		}
		commit.ObjectIDs = append(commit.ObjectIDs, objectID)
	}
	if err := rows.Err(); err != nil {
		return engine.Commit{}, fmt.Errorf("iterate commit objects: %w", err)
	}

	return commit, nil
}
