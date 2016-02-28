
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
-- Move the relationship between campaigns and smtp to campaigns
ALTER TABLE campaigns ADD COLUMN "smtp_id" bigint;
UPDATE campaigns 
	SET smtp_id = (
		SELECT smtp.smtp_id 
		FROM smtp,campaigns 
		WHERE campaigns.id=smtp.campaign_id
	)
;
-- Add the appropriate user_id to each smtp record
ALTER TABLE smtp ADD COLUMN "user_id" bigint;
UPDATE smtp 
        SET user_id = (
		SELECT campaigns.user_id 
		FROM smtp,campaigns 
		WHERE smtp.smtp_id=campaigns.smtp_id
	)
;
-- Create a new table to store smtp records
ALTER TABLE smtp RENAME TO smtp_old;
CREATE TABLE smtp(
	id integer primary key autoincrement,
	user_id bigint,
	interface_type varchar(255),
	name varchar(255),
	host varchar(255),
	username varchar(255),
	password varchar(255),
	from_address varchar(255),
	modified_date datetime default CURRENT_TIMESTAMP,
	ignore_cert_errors BOOLEAN
);
-- Import existing smtp records into new format and drop the old table
INSERT INTO smtp (id,user_id,interface_type,name,host,username,from_address,ignore_cert_errors)
	SELECT smtp_id,user_id,'SMTP',
	'Imported campaign via ' || COALESCE(host,'') || ' from ' || COALESCE(from_address,''),
	host,username,from_address,ignore_cert_errors
	FROM smtp_old
	-- Prevent insertion of duplicate records
	WHERE smtp_id IN (
		SELECT smtp_id
		FROM smtp_old
		GROUP BY user_id,host,from_address
	)
;
DROP TABLE smtp_old;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

