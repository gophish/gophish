
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `campaigns` ADD COLUMN character_set VARCHAR(255);
ALTER TABLE `campaigns` ADD COLUMN r_id_length INTEGER;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

