BEGIN;

-- Create users table
-- external_id is the primary identifier from the identity provider (WorkOS, Auth0, etc.)
-- This serves as the canonical user ID in the system
CREATE TABLE IF NOT EXISTS users (
    external_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    first_name TEXT,
    last_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
