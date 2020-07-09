-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "reported" ("id" integer primary key autoincrement, "user_id" integer ,"reported_by_name" varchar(255), "reported_by_enaio" varchar(255) "reported_time" datetime, "reported_html" varchar(255), "reported_text" varchar(255), "reported_subject" varchar(255),"imap_uid" varchar(255), "status" varchar(255), "notes" varchar(255));

CREATE TABLE IF NOT EXISTS "reported_attachments" ("id" integer primary key autoincrement, "rid" integer, "filename" varchar(255), "header" varchar(255), "size" integer, "content" varchar(255));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "reported";
DROP TABLE "reported_attachments";
