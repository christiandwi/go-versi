# go-versi

`go-versi` is a small Go foundation for learning how a repository/versioning
engine works.

Current features:

- create and get a repository
- create and get an object inside a repository
- store a SHA-256 content hash for each object
- create and get a commit
- link each new commit to its previous commit
- set and get a ref, such as `main`
- list commit history from a ref
- read the current snapshot from a ref
- avoid creating a new object, commit, or ref update when the file paths and content did not change

Current flow:

```text
repository -> object -> commit <- ref(main)
                         |
                      parent
```

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

The demo:

- creates or loads `repo-1`
- checks whether `README.md` changed compared to `main`
- creates a `README.md` object only when content changed
- creates a commit only when content changed
- links the new commit to the previous `main` commit
- sets `main` only when content changed
- reads the current ref and commit back from PostgreSQL
- prints commit history from newest to oldest
- prints the current snapshot files

If `main` already points to the same file paths and content, the demo prints
`ref unchanged` without creating a new object or commit.

## Docker

Run the test suite in Docker:

```sh
docker compose run --rm app
```

The local database URL is:

```text
postgres://goversi:goversi@localhost:5439/goversi?sslmode=disable
```
