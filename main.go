package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"go-versi/internal/engine"
	pgstore "go-versi/internal/postgres"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://goversi:goversi@localhost:5439/goversi?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	store := pgstore.New(db)
	app := engine.New(store)

	repoName := "repo-1"
	repo, err := app.CreateRepository(ctx, repoName)
	if errors.Is(err, engine.ErrConflict) {
		repo, err = app.GetRepository(ctx, engine.HashRepository(repoName))
	}
	if err != nil {
		panic(err)
	}

	found, err := app.GetRepository(ctx, repo.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("repository:", repo.ID, repo.Name)
	fmt.Println("found repository:", found.ID, found.Name)

	commit, ref, err := app.CommitToRef(ctx, repo.ID, "main", []engine.CommitChange{
		{
			Path: "README.md",
			Data: []byte("hello"),
		},
	}, "third commit")
	if errors.Is(err, engine.ErrNoChanges) {
		ref, err = app.GetRef(ctx, repo.ID, "main")
		fmt.Println("ref unchanged:", ref.Name, ref.CommitID)
	}
	if err != nil {
		panic(err)
	}
	if commit.ID != "" {
		fmt.Println("created commit:", commit.ID, commit.Message)
		fmt.Println("set ref:", ref.Name, ref.CommitID)
	}

	foundRef, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		panic(err)
	}

	fmt.Println("found ref:", foundRef.Name, foundRef.CommitID)

	foundCommit, err := app.GetCommit(ctx, foundRef.CommitID)
	if err != nil {
		panic(err)
	}

	fmt.Println("found commit:", foundCommit.ID, foundCommit.Message)

	commits, err := app.Log(ctx, repo.ID, "main")
	if err != nil {
		panic(err)
	}

	fmt.Println("log:")
	for _, commit := range commits {
		fmt.Println("-", commit.ID, commit.Message)
	}

	objects, err := app.GetSnapshot(ctx, repo.ID, "main")
	if err != nil {
		panic(err)
	}

	fmt.Println("snapshot:")
	for _, object := range objects {
		fmt.Println("-", object.Path, string(object.Data))
	}
}
