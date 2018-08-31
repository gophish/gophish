-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE campaigns ADD COLUMN send_delay integer;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
