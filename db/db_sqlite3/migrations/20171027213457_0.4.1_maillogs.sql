
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "mail_logs" (
    "id" integer primary key autoincrement,
    "campaign_id" integer,
    "user_id" integer,
    "send_date" datetime,
    "send_attempt" integer,
    "r_id" varchar(255),
    "processing" boolean);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "mail_logs"
