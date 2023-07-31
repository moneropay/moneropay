CREATE TABLE IF NOT EXISTS last_block_height (
    height INTEGER NOT NULL DEFAULT 0
);
INSERT OR IGNORE INTO last_block_height (height)
SELECT value FROM metadata WHERE key = 'last_height';
DROP TABLE IF EXISTS metadata;
-- Again, SQLite doesn't support 'alter table rename column'.
CREATE TABLE IF NOT EXISTS subaddresses_new (
    address_index INTEGER PRIMARY KEY,
    address CHARACTER(95) UNIQUE NOT NULL
);
INSERT INTO subaddresses_new (address_index, address)
SELECT `index`, address FROM subaddresses;
DROP TABLE subaddresses;
ALTER TABLE subaddresses_new RENAME TO subaddresses;
