BEGIN;
ALTER TABLE IF EXISTS receivers ADD COLUMN IF NOT EXISTS received_amount bigint DEFAULT 0,
ADD COLUMN IF NOT EXISTS creation_height bigint;
COMMIT;
