-- Migration script to update client_traffics table
-- This script changes the unique constraint on email to a composite unique constraint on (inbound_id, email)
-- This allows the same email to exist in different inbounds

-- Step 1: Drop the old unique constraint on email column
-- SQLite doesn't support dropping constraints directly, so we need to recreate the table
BEGIN TRANSACTION;

-- Create a temporary table with the new structure
CREATE TABLE IF NOT EXISTS client_traffics_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inbound_id INTEGER NOT NULL,
    enable INTEGER NOT NULL,
    email TEXT NOT NULL,
    up INTEGER NOT NULL,
    down INTEGER NOT NULL,
    all_time INTEGER NOT NULL,
    expiry_time INTEGER NOT NULL,
    total INTEGER NOT NULL,
    reset INTEGER NOT NULL DEFAULT 0,
    last_online INTEGER NOT NULL DEFAULT 0,
    UNIQUE(inbound_id, email)
);

-- Copy data from the old table to the new table
-- Use COALESCE to set inbound_id to 1 if it's NULL or 0
INSERT INTO client_traffics_new (id, inbound_id, enable, email, up, down, all_time, expiry_time, total, reset, last_online)
SELECT id, COALESCE(NULLIF(inbound_id, 0), 1), enable, email, up, down, all_time, expiry_time, total, reset, last_online
FROM client_traffics;

-- Drop the old table
DROP TABLE client_traffics;

-- Rename the new table to the original name
ALTER TABLE client_traffics_new RENAME TO client_traffics;

-- Create index on inbound_id and email for better query performance
CREATE INDEX IF NOT EXISTS idx_client_traffics_inbound_email ON client_traffics(inbound_id, email);

COMMIT;

-- Note: After running this migration, you may need to restart the application
-- to ensure the changes take effect properly.
