-- No 'alter table drop column', do the new table thing again.
CREATE TABLE IF NOT EXISTS receivers_new (
    subaddress_index INTEGER PRIMARY KEY REFERENCES subaddresses ON DELETE CASCADE,
    expected_amount INTEGER NOT NULL CHECK (expected_amount >= 0),
    description VARCHAR(1024),
    callback_url VARCHAR(2048) NOT NULL,
    created_at TIMESTAMP
);

INSERT INTO receivers_new (subaddress_index, expected_amount, description, callback_url, created_at)
SELECT subaddress_index, expected_amount, description, callback_url, created_at
FROM receivers;

DROP TABLE receivers;
ALTER TABLE receivers_new RENAME TO receivers;