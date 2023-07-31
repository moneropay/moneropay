CREATE TABLE IF NOT EXISTS metadata (
    key TEXT UNIQUE NOT NULL,
    value INTEGER NOT NULL
);
INSERT INTO metadata (key, value)
SELECT 'last_height', (SELECT height FROM last_block_height LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM metadata WHERE key = 'last_height');
DROP TABLE IF EXISTS last_block_height;
-- SQLite doesn't support 'alter table rename column', so create a new table and copy the data with the new column name.
CREATE TABLE IF NOT EXISTS subaddresses_new (
    index INTEGER PRIMARY KEY,
    address CHARACTER(95) UNIQUE NOT NULL
);

INSERT INTO subaddresses_new (index, address)
SELECT address_index, address FROM subaddresses;

DROP TABLE subaddresses;
ALTER TABLE subaddresses_new RENAME TO subaddresses;