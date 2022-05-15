BEGIN;
CREATE TABLE IF NOT EXISTS last_block_height (
	height bigint NOT NULL DEFAULT 0
);
DO $$
BEGIN
    IF EXISTS
        (SELECT 1 FROM information_schema.tables WHERE table_name = 'metadata')
    THEN
        INSERT INTO last_block_height (height) VALUES ((SELECT value FROM metadata WHERE key = 'last_height'));
    ELSE
        INSERT INTO last_block_height (height) VALUES (0);
    END IF;
END;
$$;
DROP TABLE IF EXISTS metadata;
DO $$
BEGIN
    IF EXISTS
        (SELECT 1 FROM information_schema.tables WHERE table_name = 'subaddresses')
    THEN
	ALTER TABLE subaddresses RENAME COLUMN index TO address_index;
    END IF;
END;
$$;
CREATE TABLE IF NOT EXISTS subaddresses (
	address_index bigint PRIMARY KEY,
	address character(95) UNIQUE NOT NULL,
	used_until bigint
);
CREATE TABLE IF NOT EXISTS receivers (
	subaddress_index bigint PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
	expected_amount bigint NOT NULL CHECK (expected_amount >= 0),
	description character varying(1024),
	callback_url character varying(2048) NOT NULL,
	created_at timestamp
);
COMMIT;
