-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE results ADD COLUMN url_invalidated BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE results DROP COLUMN url_invalidated; 