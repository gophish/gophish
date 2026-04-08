
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username VARCHAR(255) NOT NULL UNIQUE, hash VARCHAR(255), api_key VARCHAR(255) NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS templates (id SERIAL PRIMARY KEY, user_id BIGINT, name VARCHAR(255), subject VARCHAR(255), text TEXT, html TEXT, modified_date TIMESTAMP);
CREATE TABLE IF NOT EXISTS targets (id SERIAL PRIMARY KEY, first_name VARCHAR(255), last_name VARCHAR(255), email VARCHAR(255), position VARCHAR(255));
CREATE TABLE IF NOT EXISTS smtp (smtp_id SERIAL PRIMARY KEY, campaign_id BIGINT, host VARCHAR(255), username VARCHAR(255), from_address VARCHAR(255));
CREATE TABLE IF NOT EXISTS results (id SERIAL PRIMARY KEY, campaign_id BIGINT, user_id BIGINT, r_id VARCHAR(255), email VARCHAR(255), first_name VARCHAR(255), last_name VARCHAR(255), status VARCHAR(255) NOT NULL, ip VARCHAR(255), latitude REAL, longitude REAL);
CREATE TABLE IF NOT EXISTS pages (id SERIAL PRIMARY KEY, user_id BIGINT, name VARCHAR(255), html TEXT, modified_date TIMESTAMP);
CREATE TABLE IF NOT EXISTS groups (id SERIAL PRIMARY KEY, user_id BIGINT, name VARCHAR(255), modified_date TIMESTAMP);
CREATE TABLE IF NOT EXISTS group_targets (group_id BIGINT, target_id BIGINT);
CREATE TABLE IF NOT EXISTS events (id SERIAL PRIMARY KEY, campaign_id BIGINT, email VARCHAR(255), time TIMESTAMP, message VARCHAR(255));
CREATE TABLE IF NOT EXISTS campaigns (id SERIAL PRIMARY KEY, user_id BIGINT, name VARCHAR(255) NOT NULL, created_date TIMESTAMP, completed_date TIMESTAMP, template_id BIGINT, page_id BIGINT, status VARCHAR(255), url VARCHAR(255));
CREATE TABLE IF NOT EXISTS attachments (id SERIAL PRIMARY KEY, template_id BIGINT, content TEXT, type VARCHAR(255), name VARCHAR(255));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS group_targets;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS results;
DROP TABLE IF EXISTS smtp;
DROP TABLE IF EXISTS targets;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS users;
