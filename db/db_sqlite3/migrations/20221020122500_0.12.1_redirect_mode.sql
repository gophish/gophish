-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `pages` ADD COLUMN redirect_html TEXT;
ALTER TABLE `pages` ADD COLUMN redirect_mode TEXT;