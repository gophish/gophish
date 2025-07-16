
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE results ADD COLUMN send_date DATETIME;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

