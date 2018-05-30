-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE campaigns ADD COLUMN send_dalay Integer;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
