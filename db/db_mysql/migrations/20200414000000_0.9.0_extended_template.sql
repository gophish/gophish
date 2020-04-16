-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `targets` ADD COLUMN extended_template BLOB;
ALTER TABLE `email_requests` ADD COLUMN extended_template BLOB;
ALTER TABLE `results` ADD COLUMN extended_template BLOB;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
