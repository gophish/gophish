-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE `pages` CHANGE `html` `html` LONGTEXT CHARACTER SET latin1 COLLATE latin1_swedish_ci NULL DEFAULT NULL;
ALTER TABLE `pages` ADD `public` BOOLEAN NOT NULL AFTER `redirect_url`;
ALTER TABLE `templates` ADD `public` BOOLEAN NOT NULL AFTER `modified_date`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
