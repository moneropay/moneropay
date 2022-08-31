BEGIN;
ALTER TABLE IF EXISTS subaddresses ADD CONSTRAINT subaddresses_address_check CHECK (LENGTH (address) = 95);
CREATE TABLE IF NOT EXISTS failed_callbacks (
	uid serial PRIMARY KEY,
	subaddress_index bigint REFERENCES subaddresses ON DELETE CASCADE,
	request_body text NOT NULL,
	attempts smallint DEFAULT 1,
	next_retry timestamp with time zone NOT NULL
);
COMMIT;
