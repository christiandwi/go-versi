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

func TestCreateObject(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	obj, err := app.CreateObject(ctx, repo.ID, "README.md", []byte("hello"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}

	if obj.ID == "" {
		t.Fatal("object id is empty")
	}
	if obj.RepositoryID != repo.ID {
		t.Fatalf("repository id = %q, want %q", obj.RepositoryID, repo.ID)
	}
	if obj.Path != "README.md" {
		t.Fatalf("object path = %q, want %q", obj.Path, "README.md")
	}
	if string(obj.Data) != "hello" {
		t.Fatalf("object data = %q, want %q", obj.Data, "hello")
	}
}

func TestGetObject(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}
	created, err := app.CreateObject(ctx, repo.ID, "README.md", []byte("hello"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}

	found, err := app.GetObject(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetObject() error = %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("object id = %q, want %q", found.ID, created.ID)
	}
	if found.RepositoryID != repo.ID {
		t.Fatalf("repository id = %q, want %q", found.RepositoryID, repo.ID)
	}
	if found.Path != created.Path {
		t.Fatalf("object path = %q, want %q", found.Path, created.Path)
	}
	if string(found.Data) != string(created.Data) {
		t.Fatalf("object data = %q, want %q", found.Data, created.Data)
	}
}

func TestCreateObjectRequiresRepository(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	_, err := app.CreateObject(ctx, "missing", "README.md", []byte("hello"))
	if !errors.Is(err, engine.ErrNotFound) {
		t.Fatalf("CreateObject() error = %v, want ErrNotFound", err)
	}
}

func TestCreateObjectRequiresPath(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	_, err = app.CreateObject(ctx, repo.ID, " ", []byte("hello"))
	if !errors.Is(err, engine.ErrValidation) {
		t.Fatalf("CreateObject() error = %v, want ErrValidation", err)
	}
}

func TestCreateCommit(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}
	obj, err := app.CreateObject(ctx, repo.ID, "README.md", []byte("hello"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}

	commit, err := app.CreateCommit(ctx, repo.ID, []engine.ObjectID{obj.ID}, "initial commit")
	if err != nil {
		t.Fatalf("CreateCommit() error = %v", err)
	}

	if commit.ID == "" {
		t.Fatal("commit id is empty")
	}
	if commit.RepositoryID != repo.ID {
		t.Fatalf("repository id = %q, want %q", commit.RepositoryID, repo.ID)
	}
	if commit.Message != "initial commit" {
		t.Fatalf("message = %q, want %q", commit.Message, "initial commit")
	}
	if len(commit.ObjectIDs) != 1 || commit.ObjectIDs[0] != obj.ID {
		t.Fatalf("object ids = %v, want [%s]", commit.ObjectIDs, obj.ID)
	}
}

func TestGetCommit(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}
	obj, err := app.CreateObject(ctx, repo.ID, "README.md", []byte("hello"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}
	created, err := app.CreateCommit(ctx, repo.ID, []engine.ObjectID{obj.ID}, "initial commit")
	if err != nil {
		t.Fatalf("CreateCommit() error = %v", err)
	}

	found, err := app.GetCommit(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetCommit() error = %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("commit id = %q, want %q", found.ID, created.ID)
	}
	if found.RepositoryID != repo.ID {
		t.Fatalf("repository id = %q, want %q", found.RepositoryID, repo.ID)
	}
	if found.Message != created.Message {
		t.Fatalf("message = %q, want %q", found.Message, created.Message)
	}
	if len(found.ObjectIDs) != 1 || found.ObjectIDs[0] != obj.ID {
		t.Fatalf("object ids = %v, want [%s]", found.ObjectIDs, obj.ID)
	}
}

func TestCreateCommitRequiresObject(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	_, err = app.CreateCommit(ctx, repo.ID, nil, "initial commit")
	if !errors.Is(err, engine.ErrValidation) {
		t.Fatalf("CreateCommit() error = %v, want ErrValidation", err)
	}
}

func TestCreateCommitRejectsObjectFromDifferentRepository(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repoA, err := app.CreateRepository(ctx, uniqueRepositoryName(t)+"-a")
	if err != nil {
		t.Fatalf("CreateRepository(repoA) error = %v", err)
	}
	repoB, err := app.CreateRepository(ctx, uniqueRepositoryName(t)+"-b")
	if err != nil {
		t.Fatalf("CreateRepository(repoB) error = %v", err)
	}
	obj, err := app.CreateObject(ctx, repoA.ID, "README.md", []byte("hello"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}

	_, err = app.CreateCommit(ctx, repoB.ID, []engine.ObjectID{obj.ID}, "initial commit")
	if !errors.Is(err, engine.ErrValidation) {
		t.Fatalf("CreateCommit() error = %v, want ErrValidation", err)
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
