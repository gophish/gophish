
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE smtp ADD COLUMN send_rate INTEGER DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
