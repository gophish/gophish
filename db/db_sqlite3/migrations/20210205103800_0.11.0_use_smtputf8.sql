
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE smtp ADD COLUMN use_smtputf8 BOOLEAN;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

