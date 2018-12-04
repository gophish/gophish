-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE `templates` ADD `tag` INT(11) NOT NULL AFTER `rating`;
ALTER TABLE `pages` ADD `tag` INT(11) NOT NULL AFTER `public`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
