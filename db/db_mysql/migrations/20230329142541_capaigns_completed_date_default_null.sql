-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- Check if there are any rows where completed_date is '0000-00-00 00:00:00' or NULL
UPDATE campaigns
SET completed_date = NULL
WHERE completed_date = STR_TO_DATE('0000-00-00 00:00:00', '%Y-%m-%d %H:%i:%s') OR completed_date IS NULL;

-- Check if there are any rows where send_by_date is '0000-00-00 00:00:00' or NULL
UPDATE campaigns
SET send_by_date = NULL
WHERE send_by_date = STR_TO_DATE('0000-00-00 00:00:00', '%Y-%m-%d %H:%i:%s') OR send_by_date IS NULL;

-- Check if there are any rows where launch_date is '0000-00-00 00:00:00' or NULL
UPDATE campaigns
SET launch_date = NULL
WHERE launch_date = STR_TO_DATE('0000-00-00 00:00:00', '%Y-%m-%d %H:%i:%s') OR launch_date IS NULL;

-- Alter the table to modify column definitions and set default values to NULL
ALTER TABLE campaigns
MODIFY COLUMN completed_date datetime DEFAULT NULL,
MODIFY COLUMN send_by_date datetime DEFAULT NULL,
MODIFY COLUMN launch_date datetime DEFAULT NULL;

ALTER TABLE users MODIFY COLUMN last_login datetime DEFAULT NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE campaigns
MODIFY COLUMN completed_date datetime,
MODIFY COLUMN send_by_date datetime,
MODIFY COLUMN launch_date datetime;


ALTER TABLE users MODIFY COLUMN last_login datetime;
