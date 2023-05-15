-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
PRAGMA foreign_keys = OFF;

CREATE TEMPORARY TABLE campaigns_backup(id, user_id, name, created_date, completed_date, template_id, page_id, status, url, smtp_id, launch_date, send_by_date);
INSERT INTO campaigns_backup SELECT id, user_id, name, created_date,
    CASE WHEN completed_date = '0000-00-00 00:00:00' THEN NULL ELSE completed_date END,
    template_id, page_id, status, url, smtp_id,
    CASE WHEN launch_date = '0000-00-00 00:00:00' THEN NULL ELSE launch_date END,
    CASE WHEN send_by_date = '0000-00-00 00:00:00' THEN NULL ELSE send_by_date END
FROM campaigns;
DROP TABLE campaigns;

CREATE TABLE campaigns (id integer primary key autoincrement, user_id integer, name text NOT NULL, created_date datetime, completed_date datetime DEFAULT NULL, template_id integer, page_id integer, status text, url text, smtp_id integer, launch_date datetime DEFAULT NULL, send_by_date datetime DEFAULT NULL);
INSERT INTO campaigns SELECT * FROM campaigns_backup;
DROP TABLE campaigns_backup;

PRAGMA foreign_keys = ON;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
PRAGMA foreign_keys = OFF;

CREATE TEMPORARY TABLE campaigns_backup(id, user_id, name, created_date, completed_date, template_id, page_id, status, url, smtp_id, launch_date, send_by_date);
INSERT INTO campaigns_backup SELECT id, user_id, name, created_date, completed_date, template_id, page_id, status, url, smtp_id, launch_date, send_by_date FROM campaigns;
DROP TABLE campaigns;

CREATE TABLE campaigns (id integer primary key autoincrement, user_id integer, name text NOT NULL, created_date datetime, completed_date datetime, template_id integer, page_id integer, status text, url text, smtp_id integer, launch_date datetime, send_by_date datetime);
INSERT INTO campaigns SELECT * FROM campaigns_backup;
DROP TABLE campaigns_backup;

PRAGMA foreign_keys = ON;   
