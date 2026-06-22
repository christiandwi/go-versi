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

	obj, err := app.CreateObject(ctx, repo.ID, "README.md", []byte("hello"))
	if err != nil {
		panic(err)
	}

	fmt.Println("created object:", obj.ID, obj.Path)

	foundObj, err := app.GetObject(ctx, obj.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("found object:", foundObj.ID, foundObj.Path)

	commit, err := app.CreateCommit(ctx, repo.ID, []engine.ObjectID{obj.ID}, "initial commit")
	if err != nil {
		panic(err)
	}

	fmt.Println("created commit:", commit.ID, commit.Message)

	foundCommit, err := app.GetCommit(ctx, commit.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("found commit:", foundCommit.ID, foundCommit.Message)

	ref, err := app.SetRef(ctx, repo.ID, "main", commit.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("set ref:", ref.Name, ref.CommitID)

	foundRef, err := app.GetRef(ctx, repo.ID, "main")
	if err != nil {
		panic(err)
	}

	fmt.Println("found ref:", foundRef.Name, foundRef.CommitID)
}
