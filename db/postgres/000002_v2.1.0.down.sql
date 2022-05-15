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
CREATE TABLE IF NOT EXISTS subaddresses (
	index bigint PRIMARY KEY,
	address character(95) UNIQUE NOT NULL,
	used_until bigint
);
CREATE TABLE IF NOT EXISTS receivers (
	subaddress_index bigint PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
	expected_amount bigint NOT NULL CHECK (expected_amount >= 0),
	description character varying(1024),
	callback_url character varying(2048) NOT NULL,
	created_at timestamp with time zone
);
COMMIT;
