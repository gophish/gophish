-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `users` ADD `partner` INT NOT NULL AFTER `email`;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
