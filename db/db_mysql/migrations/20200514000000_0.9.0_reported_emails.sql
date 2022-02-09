-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS `reported` (id bigint primary key auto_increment, user_id bigint, reported_by_name varchar(255), reported_by_email varchar(255), reported_time datetime, reported_html text, reported_text text, reported_subject text, imap_uid varchar(255), status varchar(255), notes text);

CREATE TABLE IF NOT EXISTS `reported_attachments` (id bigint primary key auto_increment, rid bigint, filename varchar(255), header text, size bigint, content text);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE `reported`;
DROP TABLE `reported_attachments`;