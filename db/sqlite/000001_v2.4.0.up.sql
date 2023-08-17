CREATE TABLE IF NOT EXISTS last_block_height (
	height INTEGER NOT NULL DEFAULT 0
);

INSERT OR IGNORE INTO last_block_height (height) VALUES (0);/* DEFAULT VALUES WHERE NOT EXISTS (SELECT 1 FROM last_block_height);*/

CREATE TABLE IF NOT EXISTS subaddresses (
    address_index INTEGER PRIMARY KEY,
    address TEXT UNIQUE NOT NULL CHECK (LENGTH(address) = 95)
);

CREATE TABLE IF NOT EXISTS receivers (
    subaddress_index INTEGER PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
    expected_amount INTEGER NOT NULL CHECK (expected_amount >= 0),
    description VARCHAR(1024),
    callback_url VARCHAR(2048) NOT NULL,
    created_at TIMESTAMP,
    received_amount INTEGER DEFAULT 0,
    creation_height INTEGER
);
