# go-versi

`go-versi` is a small Go foundation for a repository engine.

For now, it does four things:

- create a repository
- get a repository by ID
- create an object inside a repository
- get an object by ID
- create a commit
- get a commit by ID

This keeps the first step easy to understand.

## Verify

Start PostgreSQL first:

```sh
docker compose up -d postgres
```

Then run tests:

```sh
go test ./...
```

## Run

Start PostgreSQL:

```sh
docker compose up -d postgres
```

Run the demo:

```sh
go run .
```

The demo creates `repo-1`, reads it back from PostgreSQL, creates a
`README.md` object, then reads that object back too, then commit and get the commit.
If `repo-1` already exists, it reads the existing repository instead.

## Docker

Run the test suite in Docker:

```sh
docker compose run --rm app
```

The database URL is:

```text
postgres://goversi:goversi@localhost:5439/goversi?sslmode=disable
```
