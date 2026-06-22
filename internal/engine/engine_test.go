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
	if obj.ContentHash != engine.HashObjectContent([]byte("hello")) {
		t.Fatalf("object content hash = %q, want %q", obj.ContentHash, engine.HashObjectContent([]byte("hello")))
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
	if found.ContentHash != created.ContentHash {
		t.Fatalf("object content hash = %q, want %q", found.ContentHash, created.ContentHash)
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

func TestSetRef(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, commit := createRepositoryObjectCommit(t, ctx, app)

	ref, err := app.SetRef(ctx, repo.ID, "main", commit.ID)
	if err != nil {
		t.Fatalf("SetRef() error = %v", err)
	}

	if ref.RepositoryID != repo.ID {
		t.Fatalf("repository id = %q, want %q", ref.RepositoryID, repo.ID)
	}
	if ref.Name != "main" {
		t.Fatalf("ref name = %q, want %q", ref.Name, "main")
	}
	if ref.CommitID != commit.ID {
		t.Fatalf("commit id = %q, want %q", ref.CommitID, commit.ID)
	}
}

func TestGetRef(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, commit := createRepositoryObjectCommit(t, ctx, app)
	created, err := app.SetRef(ctx, repo.ID, "main", commit.ID)
	if err != nil {
		t.Fatalf("SetRef() error = %v", err)
	}

	found, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}

	if found.RepositoryID != created.RepositoryID {
		t.Fatalf("repository id = %q, want %q", found.RepositoryID, created.RepositoryID)
	}
	if found.Name != created.Name {
		t.Fatalf("ref name = %q, want %q", found.Name, created.Name)
	}
	if found.CommitID != created.CommitID {
		t.Fatalf("commit id = %q, want %q", found.CommitID, created.CommitID)
	}
}

func TestSetRefMovesExistingRef(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, first := createRepositoryObjectCommit(t, ctx, app)
	secondObject, err := app.CreateObject(ctx, repo.ID, "main.go", []byte("package main"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}
	second, err := app.CreateCommit(ctx, repo.ID, []engine.ObjectID{secondObject.ID}, "second commit")
	if err != nil {
		t.Fatalf("CreateCommit() error = %v", err)
	}

	if _, err := app.SetRef(ctx, repo.ID, "main", first.ID); err != nil {
		t.Fatalf("SetRef(first) error = %v", err)
	}
	if _, err := app.SetRef(ctx, repo.ID, "main", second.ID); err != nil {
		t.Fatalf("SetRef(second) error = %v", err)
	}

	found, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}

	if found.CommitID != second.ID {
		t.Fatalf("commit id = %q, want %q", found.CommitID, second.ID)
	}
}

func TestSetRefRejectsNoChanges(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, first := createRepositoryObjectCommit(t, ctx, app)
	if _, err := app.SetRef(ctx, repo.ID, "main", first.ID); err != nil {
		t.Fatalf("SetRef(first) error = %v", err)
	}

	sameObject, err := app.CreateObject(ctx, repo.ID, "README.md", []byte("hello"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}
	sameCommit, err := app.CreateCommit(ctx, repo.ID, []engine.ObjectID{sameObject.ID}, "same content")
	if err != nil {
		t.Fatalf("CreateCommit() error = %v", err)
	}

	_, err = app.SetRef(ctx, repo.ID, "main", sameCommit.ID)
	if !errors.Is(err, engine.ErrNoChanges) {
		t.Fatalf("SetRef(sameCommit) error = %v, want ErrNoChanges", err)
	}

	found, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}
	if found.CommitID != first.ID {
		t.Fatalf("commit id = %q, want %q", found.CommitID, first.ID)
	}
}

func TestSetRefAllowsChangedContent(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, first := createRepositoryObjectCommit(t, ctx, app)
	if _, err := app.SetRef(ctx, repo.ID, "main", first.ID); err != nil {
		t.Fatalf("SetRef(first) error = %v", err)
	}

	changedObject, err := app.CreateObject(ctx, repo.ID, "README.md", []byte("hello v2"))
	if err != nil {
		t.Fatalf("CreateObject() error = %v", err)
	}
	changedCommit, err := app.CreateCommit(ctx, repo.ID, []engine.ObjectID{changedObject.ID}, "changed content")
	if err != nil {
		t.Fatalf("CreateCommit() error = %v", err)
	}

	if _, err := app.SetRef(ctx, repo.ID, "main", changedCommit.ID); err != nil {
		t.Fatalf("SetRef(changedCommit) error = %v", err)
	}

	found, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}
	if found.CommitID != changedCommit.ID {
		t.Fatalf("commit id = %q, want %q", found.CommitID, changedCommit.ID)
	}
}

func TestSetRefRejectsCommitFromDifferentRepository(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repoA, commitA := createRepositoryObjectCommit(t, ctx, app)
	repoB, err := app.CreateRepository(ctx, uniqueRepositoryName(t)+"-b")
	if err != nil {
		t.Fatalf("CreateRepository(repoB) error = %v", err)
	}

	_, err = app.SetRef(ctx, repoB.ID, "main", commitA.ID)
	if !errors.Is(err, engine.ErrValidation) {
		t.Fatalf("SetRef() error = %v, want ErrValidation", err)
	}
	if repoA.ID == repoB.ID {
		t.Fatal("test setup created duplicate repositories")
	}
}

func TestCommitToRefCreatesCommitAndRef(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	commit, ref, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello")},
	}, "initial commit")
	if err != nil {
		t.Fatalf("CommitToRef() error = %v", err)
	}

	if commit.ID == "" {
		t.Fatal("commit id is empty")
	}
	if commit.ParentID != nil {
		t.Fatalf("parent id = %q, want nil", *commit.ParentID)
	}
	if ref.CommitID != commit.ID {
		t.Fatalf("ref commit id = %q, want %q", ref.CommitID, commit.ID)
	}

	found, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}
	if found.CommitID != commit.ID {
		t.Fatalf("found ref commit id = %q, want %q", found.CommitID, commit.ID)
	}
}

func TestCommitToRefRejectsNoChangesBeforeCreatingCommit(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	first, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello")},
	}, "initial commit")
	if err != nil {
		t.Fatalf("CommitToRef(first) error = %v", err)
	}

	second, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello")},
	}, "same content")
	if !errors.Is(err, engine.ErrNoChanges) {
		t.Fatalf("CommitToRef(second) error = %v, want ErrNoChanges", err)
	}
	if second.ID != "" {
		t.Fatalf("second commit id = %q, want empty", second.ID)
	}

	found, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}
	if found.CommitID != first.ID {
		t.Fatalf("ref commit id = %q, want %q", found.CommitID, first.ID)
	}
}

func TestCommitToRefAllowsChangedContent(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	first, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello")},
	}, "initial commit")
	if err != nil {
		t.Fatalf("CommitToRef(first) error = %v", err)
	}

	second, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello v2")},
	}, "second commit")
	if err != nil {
		t.Fatalf("CommitToRef(second) error = %v", err)
	}
	if second.ID == "" || second.ID == first.ID {
		t.Fatalf("second commit id = %q, first = %q", second.ID, first.ID)
	}
	if second.ParentID == nil || *second.ParentID != first.ID {
		t.Fatalf("second parent id = %v, want %q", second.ParentID, first.ID)
	}

	found, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}
	if found.CommitID != second.ID {
		t.Fatalf("ref commit id = %q, want %q", found.CommitID, second.ID)
	}

	foundCommit, err := app.GetCommit(ctx, second.ID)
	if err != nil {
		t.Fatalf("GetCommit(second) error = %v", err)
	}
	if foundCommit.ParentID == nil || *foundCommit.ParentID != first.ID {
		t.Fatalf("found parent id = %v, want %q", foundCommit.ParentID, first.ID)
	}
}

func TestLogReturnsCommitsNewestToOldest(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	first, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello")},
	}, "initial commit")
	if err != nil {
		t.Fatalf("CommitToRef(first) error = %v", err)
	}
	second, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello v2")},
	}, "second commit")
	if err != nil {
		t.Fatalf("CommitToRef(second) error = %v", err)
	}

	commits, err := app.Log(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("commits len = %d, want 2", len(commits))
	}
	if commits[0].ID != second.ID {
		t.Fatalf("commits[0] = %q, want %q", commits[0].ID, second.ID)
	}
	if commits[1].ID != first.ID {
		t.Fatalf("commits[1] = %q, want %q", commits[1].ID, first.ID)
	}
}

func TestLogRequiresExistingRef(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	_, err = app.Log(ctx, repo.ID, "main")
	if !errors.Is(err, engine.ErrNotFound) {
		t.Fatalf("Log() error = %v, want ErrNotFound", err)
	}
}

func TestGetSnapshotReturnsObjectsFromCurrentRef(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	if _, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello")},
	}, "initial commit"); err != nil {
		t.Fatalf("CommitToRef(first) error = %v", err)
	}
	if _, _, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{Path: "README.md", Data: []byte("hello v2")},
		{Path: "main.go", Data: []byte("package main")},
	}, "second commit"); err != nil {
		t.Fatalf("CommitToRef(second) error = %v", err)
	}

	objects, err := app.GetSnapshot(ctx, repo.ID, "main")
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}

	if len(objects) != 2 {
		t.Fatalf("objects len = %d, want 2", len(objects))
	}

	byPath := make(map[string]string, len(objects))
	for _, obj := range objects {
		byPath[obj.Path] = string(obj.Data)
	}
	if byPath["README.md"] != "hello v2" {
		t.Fatalf("README.md = %q, want %q", byPath["README.md"], "hello v2")
	}
	if byPath["main.go"] != "package main" {
		t.Fatalf("main.go = %q, want %q", byPath["main.go"], "package main")
	}
}

func TestGetSnapshotRequiresExistingRef(t *testing.T) {
	ctx := context.Background()
	app := newTestEngine(t)

	repo, err := app.CreateRepository(ctx, uniqueRepositoryName(t))
	if err != nil {
		t.Fatalf("CreateRepository() error = %v", err)
	}

	_, err = app.GetSnapshot(ctx, repo.ID, "main")
	if !errors.Is(err, engine.ErrNotFound) {
		t.Fatalf("GetSnapshot() error = %v, want ErrNotFound", err)
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

func createRepositoryObjectCommit(t *testing.T, ctx context.Context, app *engine.Engine) (engine.Repository, engine.Commit) {
	t.Helper()

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

	return repo, commit
}
