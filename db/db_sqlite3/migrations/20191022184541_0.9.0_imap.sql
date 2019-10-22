
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "imap" ("user_id" bigint, "host" varchar(255), "port" integer, "username" varchar(255), "password" varchar(255), "modified_date" datetime default CURRENT_TIMESTAMP, "tls" BOOLEAN, "enabled" BOOLEAN, "folder" varchar(255), "restrict_domain" varchar(255), "delete_campaign" BOOLEAN, "last_login" datetime, "last_login_friendly" string);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "imap";
