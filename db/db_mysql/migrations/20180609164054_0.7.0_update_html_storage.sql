
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `templates` MODIFY html MEDIUMTEXT;
ALTER TABLE `pages` MODIFY html MEDIUMTEXT;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE `templates` MODIFY html TEXT;
ALTER TABLE `pages` MODIFY html TEXT;
