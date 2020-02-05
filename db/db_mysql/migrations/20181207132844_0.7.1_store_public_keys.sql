-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS public_keys (
	`id` integer primary key autoincrement,
	`user_id` integer NOT NULL, 
	`friendly_name` varchar(255), 
	`pub_key` blob NOT NULL
);

ALTER TABLE campaigns ADD COLUMN public_key_id integer;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE public_keys;
