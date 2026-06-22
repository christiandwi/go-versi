CREATE TABLE IF NOT EXISTS refs (
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    name TEXT NOT NULL,
    commit_id TEXT NOT NULL REFERENCES commits(id),
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (repository_id, name)
);
