-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE users ADD COLUMN customer_id INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
-- Note: Dropping columns in SQLite requires recreating the table. Rolling back
-- is not implemented here. Restore from backup if necessary.
