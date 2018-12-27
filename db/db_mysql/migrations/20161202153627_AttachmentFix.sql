
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `attachments` MODIFY content LONGTEXT;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE `attachments` MODIFY content TEXT;
