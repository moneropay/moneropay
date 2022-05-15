BEGIN;
CREATE TABLE IF NOT EXISTS metadata (
        key text UNIQUE NOT NULL,
        value bigint NOT NULL
);
INSERT INTO metadata (key, value) VALUES ('last_height', 0) ON CONFLICT DO NOTHING;
ALTER TABLE IF EXISTS subaddresses ADD CONSTRAINT subaddresses_address_check CHECK (LENGTH (address) = 95);
CREATE TABLE IF NOT EXISTS subaddresses (
        index bigint PRIMARY KEY,
        address character(95) UNIQUE NOT NULL CHECK (LENGTH (address) = 95)
);
CREATE TABLE IF NOT EXISTS receivers (
        subaddress_index bigint PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
        expected_amount bigint NOT NULL CHECK (expected_amount >= 0),
        description character varying(1024),
        callback_url character varying(2048) NOT NULL,
        created_at timestamp with time zone
);
CREATE TABLE IF NOT EXISTS failed_callbacks (
	uid serial PRIMARY KEY,
	subaddress_index bigint REFERENCES subaddresses ON DELETE CASCADE,
	request_body text NOT NULL,
	attempts smallint DEFAULT 1,
	next_retry timestamp with time zone NOT NULL
);
COMMIT;
