-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE campaigns ADD COLUMN send_by_date DATETIME;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
