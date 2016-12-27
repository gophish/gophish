
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "campaigns" ADD COLUMN "delay" integer;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

