-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- Add OAuth columns to users table
ALTER TABLE users ADD COLUMN oauth_provider VARCHAR(255);
ALTER TABLE users ADD COLUMN oauth_subject VARCHAR(255);

-- Create index for OAuth lookups
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_subject);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP INDEX idx_users_oauth;
ALTER TABLE users DROP COLUMN oauth_subject;
ALTER TABLE users DROP COLUMN oauth_provider;
