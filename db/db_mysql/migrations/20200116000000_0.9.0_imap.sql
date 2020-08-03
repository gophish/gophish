
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS `imap` (user_id bigint,host varchar(255),port int,username varchar(255),password varchar(255),modified_date datetime,tls boolean,enabled boolean,folder varchar(255),restrict_domain varchar(255),delete_reported_campaign_email boolean,last_login datetime,imap_freq int);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE `imap`;
