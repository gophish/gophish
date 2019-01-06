
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
UPDATE `campaigns` SET `created_date`=CONVERT_TZ(`created_date`, @@session.time_zone, '+00:00');
UPDATE `campaigns` SET `completed_date`=CONVERT_TZ(`completed_date`, @@session.time_zone, '+00:00');
UPDATE `campaigns` SET `launch_date`=CONVERT_TZ(`launch_date`, @@session.time_zone, '+00:00');
UPDATE `events` SET `time`=CONVERT_TZ(`time`, @@session.time_zone, '+00:00');
UPDATE `groups` SET `modified_date`=CONVERT_TZ(`modified_date`, @@session.time_zone, '+00:00');
UPDATE `templates` SET `modified_date`=CONVERT_TZ(`modified_date`, @@session.time_zone, '+00:00');
UPDATE `pages` SET `modified_date`=CONVERT_TZ(`modified_date`, @@session.time_zone, '+00:00');
UPDATE `smtp` SET `modified_date`=CONVERT_TZ(`modified_date`, @@session.time_zone, '+00:00');


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

