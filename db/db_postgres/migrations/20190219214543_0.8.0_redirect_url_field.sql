
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE pages ALTER COLUMN redirect_url TYPE TEXT;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE pages ALTER COLUMN redirect_url TYPE VARCHAR(255);
