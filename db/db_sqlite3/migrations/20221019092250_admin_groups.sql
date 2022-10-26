-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "admin_groups" (
    "id" integer primary key autoincrement,
    "name" varchar(255)
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "admin_groups";
