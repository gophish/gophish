-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `templates` ADD `rating` INT(1) NOT NULL AFTER `html`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
