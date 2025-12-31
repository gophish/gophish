-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- Create oauth_providers table
CREATE TABLE IF NOT EXISTS oauth_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    issuer_url VARCHAR(255) NOT NULL,
    scopes TEXT,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME,
    modified_at DATETIME
);

-- Create index for lookups
CREATE INDEX idx_oauth_providers_name ON oauth_providers(name);
CREATE INDEX idx_oauth_providers_enabled ON oauth_providers(enabled);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP INDEX IF EXISTS idx_oauth_providers_enabled;
DROP INDEX IF EXISTS idx_oauth_providers_name;
DROP TABLE IF EXISTS oauth_providers;
