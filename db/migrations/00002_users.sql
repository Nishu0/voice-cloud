-- +goose Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL COLLATE NOCASE UNIQUE
        CHECK (length(trim(email)) > 0 AND email = trim(email)),
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    password_hash TEXT NOT NULL CHECK (length(trim(password_hash)) > 0),
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    email_verified_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000)
);

-- Only hashes of random session/reset/verification tokens belong in SQLite.
CREATE TABLE user_sessions (
    token_hash TEXT PRIMARY KEY NOT NULL CHECK (length(token_hash) = 64 AND token_hash NOT GLOB '*[^0-9a-f]*'),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    expires_at INTEGER NOT NULL CHECK (expires_at > created_at),
    revoked_at INTEGER CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX user_sessions_user ON user_sessions(user_id);
CREATE INDEX user_sessions_expiry ON user_sessions(expires_at);

CREATE TABLE user_tokens (
    token_hash TEXT PRIMARY KEY NOT NULL CHECK (length(token_hash) = 64 AND token_hash NOT GLOB '*[^0-9a-f]*'),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose TEXT NOT NULL CHECK (purpose IN ('password_reset', 'email_verification')),
    created_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    expires_at INTEGER NOT NULL CHECK (expires_at > created_at),
    used_at INTEGER CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX user_tokens_user ON user_tokens(user_id, purpose);
CREATE INDEX user_tokens_expiry ON user_tokens(expires_at);

-- NULL preserves attribution for existing bootstrap/system keys.
ALTER TABLE api_keys ADD COLUMN created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX api_keys_creator ON api_keys(created_by_user_id);

-- +goose Down
DROP INDEX api_keys_creator;
ALTER TABLE api_keys DROP COLUMN created_by_user_id;
DROP TABLE user_tokens;
DROP TABLE user_sessions;
DROP TABLE users;
