
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE pages ADD COLUMN capture_credentials boolean;
ALTER TABLE pages ADD COLUMN capture_passwords boolean;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

