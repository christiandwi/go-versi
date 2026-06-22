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
}
