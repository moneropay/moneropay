BEGIN;
CREATE TABLE IF NOT EXISTS metadata (
	key text UNIQUE NOT NULL,
	value bigint NOT NULL
);
DO $$
BEGIN
    IF EXISTS
        (SELECT 1 FROM information_schema.tables WHERE table_name = 'last_block_height')
    THEN
        INSERT INTO metadata (key, value) VALUES ('last_height', (SELECT height FROM last_block_height LIMIT 1));
    ELSE
        INSERT INTO metadata (key, value) VALUES ('last_height', 0);
    END IF;
END;
$$;
DROP TABLE IF EXISTS last_block_height;
DO $$
BEGIN
    IF EXISTS
        (SELECT 1 FROM information_schema.tables WHERE table_name = 'subaddresses')
    THEN
	ALTER TABLE subaddresses RENAME COLUMN address_index TO index;
    END IF;
END;
$$;
COMMIT;
