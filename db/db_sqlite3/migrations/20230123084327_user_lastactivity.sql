
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `users`
ADD COLUMN `last_acitivty` DATETIME NULL;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

