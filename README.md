# go-versi

`go-versi` is a small Go foundation for a repository engine.

For now, it only does two things:

- create a repository
- get a repository by ID

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

The demo creates `repo-1`, then reads it back from PostgreSQL. If `repo-1`
already exists, it reads the existing repository instead.

## Docker

Run the test suite in Docker:

```sh
docker compose run --rm app
```

The database URL inside Compose is:

```text
postgres://goversi:goversi@postgres:5432/goversi?sslmode=disable
```

From your local machine, use:

```text
postgres://goversi:goversi@localhost:5439/goversi?sslmode=disable
```
