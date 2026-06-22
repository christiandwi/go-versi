ALTER TABLE objects
    ADD COLUMN IF NOT EXISTS content_hash TEXT;
