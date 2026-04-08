-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
UPDATE results
SET status = 'Submitted Data'
WHERE id IN (
        SELECT results.id
        FROM results
        INNER JOIN events ON results.email = events.email
            AND results.campaign_id = events.campaign_id
        WHERE results.status = 'Success'
            AND events.message = 'Submitted Data');

UPDATE results
SET status = 'Clicked Link'
WHERE id IN (
        SELECT results.id
        FROM results
        INNER JOIN events ON results.email = events.email
            AND results.campaign_id = events.campaign_id
        WHERE results.status = 'Success'
            AND events.message = 'Clicked Link');

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
