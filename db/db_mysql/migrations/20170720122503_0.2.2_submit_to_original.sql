-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE pages ADD COLUMN submit_to_original BOOLEAN;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
