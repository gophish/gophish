-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE `campaigns`  ADD `start_time` VARCHAR(8) NOT NULL  AFTER `send_by_date`,  ADD `end_time` VARCHAR(8) NOT NULL  AFTER `start_time`,  ADD `time_zone` VARCHAR(55) NOT NULL  AFTER `end_time`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
