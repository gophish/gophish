-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE tenants ADD COLUMN updated_at TIMESTAMP;
UPDATE tenants SET updated_at = created_at WHERE updated_at IS NULL;

ALTER TABLE provider_tenants ADD COLUMN updated_at TIMESTAMP;
UPDATE provider_tenants SET updated_at = created_at WHERE updated_at IS NULL;

ALTER TABLE oauth_tokens ADD COLUMN updated_at TIMESTAMP;
UPDATE oauth_tokens SET updated_at = created_at WHERE updated_at IS NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE tenants DROP COLUMN updated_at;
ALTER TABLE provider_tenants DROP COLUMN updated_at;
ALTER TABLE oauth_tokens DROP COLUMN updated_at; 