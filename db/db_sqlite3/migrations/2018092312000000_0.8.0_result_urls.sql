-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE results ADD COLUMN `url` VARCHAR(255);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
