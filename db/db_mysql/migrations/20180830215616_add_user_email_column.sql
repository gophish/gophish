-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `users` ADD `email` VARCHAR(100) NOT NULL AFTER `username`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
