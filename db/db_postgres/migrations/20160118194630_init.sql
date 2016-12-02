
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS users (id serial primary key, username text not null unique, hash text, api_key text not null unique);
CREATE TABLE IF NOT EXISTS templates (id serial primary key, user_id bigint, name text, subject text, text text, html text, modified_date timestamp);
CREATE TABLE IF NOT EXISTS targets (id serial primary key, first_name text, last_name text, email text, position text);
CREATE TABLE IF NOT EXISTS smtp (smtp_id serial primary key, campaign_id bigint, host text, username text, from_address text);
CREATE TABLE IF NOT EXISTS results (id serial primary key, campaign_id bigint, user_id bigint, r_id text, email text, first_name text, last_name text, status text not null, ip text, latitude real, longitude real);
CREATE TABLE IF NOT EXISTS pages (id serial primary key, user_id bigint, name text, html text, modified_date timestamp);
CREATE TABLE IF NOT EXISTS groups (id serial primary key, user_id bigint, name text, modified_date timestamp);
CREATE TABLE IF NOT EXISTS group_targets (group_id bigint, target_id bigint);
CREATE TABLE IF NOT EXISTS events (id serial primary key, campaign_id bigint, email text, time timestamp, message text);
CREATE TABLE IF NOT EXISTS campaigns (id serial primary key, user_id bigint, name text not null, created_date timestamp, completed_date timestamp, template_id bigint, page_id bigint, status text, url text);
CREATE TABLE IF NOT EXISTS attachments (id serial primary key, template_id bigint, content text, type text, name text);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE attachments;
DROP TABLE campaigns;
DROP TABLE events;
DROP TABLE group_targets;
DROP TABLE groups;
DROP TABLE pages;
DROP TABLE results;
DROP TABLE smtp;
DROP TABLE targets;
DROP TABLE templates;
DROP TABLE users;
