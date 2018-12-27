-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE `users` ADD UNIQUE(`email`);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
