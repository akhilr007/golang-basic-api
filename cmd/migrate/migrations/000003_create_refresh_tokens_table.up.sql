CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    token_hash TEXT NOT NULL,
    jti TEXT NOT NULL UNIQUE,
    family_id TEXT NOT NULL,

    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_refresh_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_family ON refresh_tokens(family_id);
