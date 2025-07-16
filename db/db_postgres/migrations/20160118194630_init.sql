
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY,username varchar(255) NOT NULL UNIQUE,hash varchar(255),api_key varchar(255) NOT NULL UNIQUE );
CREATE TABLE IF NOT EXISTS templates (id SERIAL PRIMARY KEY,user_id bigint,name varchar(255),subject varchar(255),text text,html text,modified_date timestamp );
CREATE TABLE IF NOT EXISTS targets (id SERIAL PRIMARY KEY,first_name varchar(255),last_name varchar(255),email varchar(255),position varchar(255) );
CREATE TABLE IF NOT EXISTS smtp (smtp_id SERIAL PRIMARY KEY,campaign_id bigint,host varchar(255),username varchar(255),from_address varchar(255) );
CREATE TABLE IF NOT EXISTS results (id SERIAL PRIMARY KEY,campaign_id bigint,user_id bigint,r_id varchar(255),email varchar(255),first_name varchar(255),last_name varchar(255),status varchar(255) NOT NULL ,ip varchar(255),latitude double precision,longitude double precision );
CREATE TABLE IF NOT EXISTS pages (id SERIAL PRIMARY KEY,user_id bigint,name varchar(255),html text,modified_date timestamp );
CREATE TABLE IF NOT EXISTS groups (id SERIAL PRIMARY KEY,user_id bigint,name varchar(255),modified_date timestamp );
CREATE TABLE IF NOT EXISTS group_targets (group_id bigint,target_id bigint );
CREATE TABLE IF NOT EXISTS events (id SERIAL PRIMARY KEY,campaign_id bigint,email varchar(255),time timestamp,message varchar(255) );
CREATE TABLE IF NOT EXISTS campaigns (id SERIAL PRIMARY KEY,user_id bigint,name varchar(255) NOT NULL ,created_date timestamp,completed_date timestamp,template_id bigint,page_id bigint,status varchar(255),url varchar(255) );
CREATE TABLE IF NOT EXISTS attachments (id SERIAL PRIMARY KEY,template_id bigint,content text,type varchar(255),name varchar(255) );

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