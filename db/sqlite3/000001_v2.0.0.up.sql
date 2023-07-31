CREATE TABLE IF NOT EXISTS metadata (
    key TEXT UNIQUE NOT NULL,
    value INTEGER NOT NULL
);
INSERT OR IGNORE INTO metadata (key, value) VALUES ('last_height', 0);
CREATE TABLE IF NOT EXISTS subaddresses (
    `index` INTEGER PRIMARY KEY,
    address TEXT UNIQUE NOT NULL CHECK (LENGTH(address) = 95)
);
CREATE TABLE IF NOT EXISTS receivers (
    subaddress_index INTEGER PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
    expected_amount INTEGER NOT NULL CHECK (expected_amount >= 0),
    description TEXT,
    callback_url TEXT NOT NULL,
    created_at TIMESTAMP
);
DROP TABLE IF EXISTS failed_callbacks;