package engine_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"go-versi/internal/engine"
	pgstore "go-versi/internal/postgres"

	_ "github.com/lib/pq"
)

func TestCreateRepository(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	name := uniqueRepositoryName(t)
	repo, err := app.CreateRepository(ctx, name)
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	if repo.ID == "" {
		t.Fatal("repository id is empty")
	}
	if repo.Name != name {
		t.Fatalf("repository name = %q, want %q", repo.Name, name)
	}
}

func TestGetRepository(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	created, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	found, err := app.GetRepository(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetRepository() error = %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("repository id = %q, want %q", found.ID, created.ID)
	}
	if found.Name != created.Name {
		t.Fatalf("repository name = %q, want %q", found.Name, created.Name)
	}
}

func TestGetRepositoryNotFound(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	_, err := app.GetRepository(ctx, "missing")
	if !errors.Is(err, engine.ErrNotFound) {
		t.Fatalf("GetRepository() error = %v, want ErrNotFound", err)
	}
}

func TestCreateRepositoryRequiresName(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	_, err := app.CreateRepository(ctx, " ")
	if !errors.Is(err, engine.ErrValidation) {
		t.Fatalf("CreateRepository() error = %v, want ErrValidation", err)
	}
}

func newTestEngine(t *testing.T) *engine.Engine {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://goversi:goversi@localhost:5439/goversi?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() error = %v", err)
	}

	return engine.New(pgstore.New(db))
}

func uniqueRepositoryName(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
}
