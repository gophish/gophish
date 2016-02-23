
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE campaigns ADD COLUMN smtp_id bigint;
UPDATE campaigns 
	SET smtp_id = (SELECT smtp.smtp_id FROM smtp) 
	WHERE campaigns.id = (
		SELECT campaigns.id
		FROM smtp,campaigns 
		WHERE smtp.campaign_id=campaigns.id
	)
;  -- sure hope the current smtp table works like I think it does
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
INSERT INTO smtp (id,interface_type,name,host,username,from_address,ignore_cert_errors)
	SELECT smtp_id,'SMTP',smtp_id,host,username,from_address,ignore_cert_errors 
	FROM smtp_old
;
UPDATE smtp 
	SET user_id = (SELECT campaigns.user_id FROM campaigns) 
	WHERE smtp.id = (
		SELECT smtp.id 
		FROM smtp,campaigns 
		WHERE smtp.id=campaigns.smtp_id
	)
;
DROP TABLE smtp_old;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

