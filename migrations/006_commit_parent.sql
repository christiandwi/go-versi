ALTER TABLE commits
    ADD COLUMN IF NOT EXISTS parent_id TEXT REFERENCES commits(id);
