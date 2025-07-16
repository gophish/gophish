
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE templates ADD COLUMN envelope_sender varchar(255);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

