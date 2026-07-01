-- up
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token TEXT PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS password_reset_tokens;
