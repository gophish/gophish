-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
UPDATE `campaigns`
SET `completed_date` = NULL
WHERE `completed_date` = '0000-00-00 00:00:00';

UPDATE `campaigns`
SET `send_by_date` = NULL
WHERE `send_by_date` = '0000-00-00 00:00:00';

UPDATE `campaigns`
SET `launch_date` = NULL
WHERE `launch_date` = '0000-00-00 00:00:00';

ALTER TABLE `campaigns`
MODIFY COLUMN `completed_date` datetime DEFAULT NULL,
MODIFY COLUMN `send_by_date` datetime DEFAULT NULL,
MODIFY COLUMN `launch_date` datetime DEFAULT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE `campaigns`
MODIFY COLUMN `completed_date` datetime,
MODIFY COLUMN `send_by_date` datetime,
MODIFY COLUMN `launch_date` datetime;
