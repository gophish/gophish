
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "campaigns" ADD COLUMN "smtp_min_delay" VARCHAR(255);
ALTER TABLE "campaigns" ADD COLUMN "smtp_max_delay" VARCHAR(255);

UPDATE "campaigns" SET "smtp_min_delay" = "1";
UPDATE "campaigns" SET "smtp_max_delay" = "10";
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

