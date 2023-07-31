ALTER TABLE subaddresses ADD CHECK (LENGTH(address) = 95);
CREATE TABLE IF NOT EXISTS failed_callbacks (
    uid INTEGER PRIMARY KEY AUTOINCREMENT,
    subaddress_index INTEGER REFERENCES subaddresses ON DELETE CASCADE,
    request_body TEXT NOT NULL,
    attempts SMALLINT DEFAULT 1,
    next_retry TIMESTAMP NOT NULL
);