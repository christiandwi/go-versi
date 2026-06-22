CREATE TABLE IF NOT EXISTS commits (
    id TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    message TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS commit_objects (
    commit_id TEXT NOT NULL REFERENCES commits(id) ON DELETE CASCADE,
    object_id TEXT NOT NULL REFERENCES objects(id) ON DELETE RESTRICT,
    position INTEGER NOT NULL,
    PRIMARY KEY (commit_id, position),
    UNIQUE (commit_id, object_id)
);
