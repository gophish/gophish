-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE campaigns ADD COLUMN cc TEXT;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE campaigns DROP COLUMN cc;
