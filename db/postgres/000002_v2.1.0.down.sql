BEGIN;
CREATE TABLE IF NOT EXISTS metadata (
	key text UNIQUE NOT NULL,
	value bigint NOT NULL
);
INSERT INTO metadata (key, value) SELECT 'last_height', (SELECT height FROM last_block_height LIMIT 1) WHERE NOT EXISTS (SELECT 1 FROM metadata);
DROP TABLE IF EXISTS last_block_height;
ALTER TABLE subaddresses RENAME COLUMN address_index TO index;
COMMIT;
