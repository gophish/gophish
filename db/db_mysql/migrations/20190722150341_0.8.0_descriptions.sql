-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE "campaigns" ADD COLUMN "description" VARCHAR(3000);
ALTER TABLE "templates" ADD COLUMN "description" VARCHAR(3000);
ALTER TABLE "pages" ADD COLUMN "description" VARCHAR(3000);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "campaigns" DROP COLUMN "description";
ALTER TABLE "templates" DROP COLUMN "description";
ALTER TABLE "pages" DROP COLUMN "description";
