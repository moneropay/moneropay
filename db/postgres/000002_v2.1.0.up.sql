BEGIN;
CREATE TABLE IF NOT EXISTS last_block_height (
	height bigint NOT NULL DEFAULT 0
);
INSERT INTO last_block_height (height) SELECT (SELECT value FROM metadata WHERE key = 'last_height') WHERE NOT EXISTS (SELECT 1 FROM last_block_height);
DROP TABLE IF EXISTS metadata;
ALTER TABLE subaddresses RENAME COLUMN index TO address_index;
COMMIT;
