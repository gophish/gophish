
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
-- Move the relationship between campaigns and smtp to campaigns
ALTER TABLE campaigns ADD COLUMN smtp_id bigint;
-- Create a new table to store smtp records
DROP TABLE smtp;
CREATE TABLE smtp(
	id serial primary key, 
	user_id bigint, 
	interface_type text, 
	name text, 
	host text, 
	username text, 
	password text, 
	from_address text, 
	modified_date timestamp, 
	ignore_cert_errors boolean
);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

