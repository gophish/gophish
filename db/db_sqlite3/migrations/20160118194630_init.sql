
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "users" ("id" integer primary key autoincrement,"username" varchar(255) NOT NULL UNIQUE,"hash" varchar(255),"api_key" varchar(255) NOT NULL UNIQUE );
CREATE TABLE IF NOT EXISTS "templates" ("id" integer primary key autoincrement,"user_id" bigint,"name" varchar(255),"subject" varchar(255),"text" varchar(255),"html" varchar(255),"modified_date" datetime );
CREATE TABLE IF NOT EXISTS "targets" ("id" integer primary key autoincrement,"first_name" varchar(255),"last_name" varchar(255),"email" varchar(255),"position" varchar(255) );
CREATE TABLE IF NOT EXISTS "smtp" ("smtp_id" integer primary key autoincrement,"campaign_id" bigint,"host" varchar(255),"username" varchar(255),"from_address" varchar(255) );
CREATE TABLE IF NOT EXISTS "results" ("id" integer primary key autoincrement,"campaign_id" bigint,"user_id" bigint,"r_id" varchar(255),"email" varchar(255),"first_name" varchar(255),"last_name" varchar(255),"status" varchar(255) NOT NULL ,"ip" varchar(255),"latitude" real,"longitude" real );
CREATE TABLE IF NOT EXISTS "pages" ("id" integer primary key autoincrement,"user_id" bigint,"name" varchar(255),"html" varchar(255),"modified_date" datetime );
CREATE TABLE IF NOT EXISTS "groups" ("id" integer primary key autoincrement,"user_id" bigint,"name" varchar(255),"modified_date" datetime );
CREATE TABLE IF NOT EXISTS "group_targets" ("group_id" bigint,"target_id" bigint );
CREATE TABLE IF NOT EXISTS "events" ("id" integer primary key autoincrement,"campaign_id" bigint,"email" varchar(255),"time" datetime,"message" varchar(255) );
CREATE TABLE IF NOT EXISTS "campaigns" ("id" integer primary key autoincrement,"user_id" bigint,"name" varchar(255) NOT NULL ,"created_date" datetime,"completed_date" datetime,"template_id" bigint,"page_id" bigint,"status" varchar(255),"url" varchar(255) );
CREATE TABLE IF NOT EXISTS "attachments" ("id" integer primary key autoincrement,"template_id" bigint,"content" varchar(255),"type" varchar(255),"name" varchar(255) );

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "attachments";
DROP TABLE "campaigns";
DROP TABLE "events";
DROP TABLE "group_targets";
DROP TABLE "groups";
DROP TABLE "pages";
DROP TABLE "results";
DROP TABLE "smtp";
DROP TABLE "targets";
DROP TABLE "templates";
DROP TABLE "users";