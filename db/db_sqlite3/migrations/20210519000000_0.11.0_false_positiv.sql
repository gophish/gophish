-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE events ADD COLUMN false_positive BOOLEAN DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
