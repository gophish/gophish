
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE campaigns ADD COLUMN smtp_id bigint;
UPDATE campaigns SET smtp_id = (SELECT smtp.smtp_id FROM smtp) WHERE campaigns.id = (SELECT smtp.smtp_id FROM smtp,campaigns WHERE smtp.campaign_id=campaigns.id);  -- sure hope the current smtp table works like I think it does

ALTER TABLE smtp RENAME TO smtp_old;
CREATE TABLE smtp(id integer primary key autoincrement,host varchar(255),username varchar(255),from_address varchar(255),modified_date datetime default CURRENT_TIMESTAMP,ignore_cert_errors BOOLEAN);
INSERT INTO smtp SELECT smtp_id AS id,host,username,from_address,'',ignore_cert_errors FROM smtp_old;
DROP TABLE smtp_old;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

