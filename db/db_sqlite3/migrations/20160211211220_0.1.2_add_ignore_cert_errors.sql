
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE smtp ADD COLUMN ignore_cert_errors BOOLEAN;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

