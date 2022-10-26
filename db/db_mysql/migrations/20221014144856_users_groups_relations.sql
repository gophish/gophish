-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "users_admin_groups" (
    "id" integer primary key autoincrement,
    "user_id" integer,
    "admin_group_id" integer
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "users_admin_groups";
