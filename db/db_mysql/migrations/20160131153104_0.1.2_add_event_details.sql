
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE events ADD COLUMN details BLOB;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

