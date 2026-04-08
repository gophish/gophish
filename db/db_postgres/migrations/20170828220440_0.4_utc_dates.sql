
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
UPDATE campaigns SET created_date = created_date AT TIME ZONE 'UTC';
UPDATE campaigns SET completed_date = completed_date AT TIME ZONE 'UTC';
UPDATE campaigns SET launch_date = launch_date AT TIME ZONE 'UTC';
UPDATE events SET time = time AT TIME ZONE 'UTC';
UPDATE groups SET modified_date = modified_date AT TIME ZONE 'UTC';
UPDATE templates SET modified_date = modified_date AT TIME ZONE 'UTC';
UPDATE pages SET modified_date = modified_date AT TIME ZONE 'UTC';
UPDATE smtp SET modified_date = modified_date AT TIME ZONE 'UTC';


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

