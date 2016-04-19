
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE campaigns RENAME TO campaigns_old;
/*
The new campaigns table will rely on a "flows" table that is basically a list of tasks to accomplish.

This will include sending emails, hosting landing pages, or possibly taking other action as needed,
including running scripts, sending HTTP requests, etc.
*/
CREATE TABLE campaigns ("id" integer primary key autoincrement, "user_id" bigint, "name" varchar(255) NOT NULL, "created_date" datetime, "completed_date" datetime, "flow_id" bigint, "status" varchar(255)); 

INSERT INTO campaigns
(
	id,
	user_id,
	name,
	created_date,
	completed_date,
	status
)
SELECT
	campaigns_old.id,
	campaigns_old.user_id,
	campaigns_old.name,
	campaigns_old.created_date,
	campaigns_old.completed_date,
	campaigns_old.status
FROM campaigns_old;

/* Create our flows table */

CREATE TABLE "flows" ("id" integer primary key autoincrement,"user_id" bigint,"previous_id" bigint,"next_id" bigint,"campaign_id" bigint, "metadata" blob,"task" varchar(255));

/* Setup our email flows */

INSERT INTO flows (
	user_id,
	campaign_id,
	task,
	metadata
)
SELECT
	campaigns_old.user_id,
	campaigns_old.id,
	"SEND_EMAIL",
	'{'                      ||
		'"smtp_id":'     || campaigns_old.smtp_id     || ',' ||
		'"template_id":' || campaigns_old.template_id ||
	'}'
FROM campaigns_old;

/* Setup our landing page flows */

INSERT INTO flows (
	user_id,
	campaign_id,
	task,
	metadata,
	previous_id
)
SELECT
	campaigns_old.user_id,
	campaigns_old.id as campaign_id,
	"LANDING_PAGE",
	'{'                    ||
		'"page_id" : ' || campaigns_old.page_id || ',' ||
		'"url" : "'    || campaigns_old.url     || '"' ||
	'}',
	flows.id as flow_id
FROM campaigns_old, flows
WHERE flows.id IN (
	SELECT id FROM flows
	WHERE campaign_id=campaigns_old.id
		AND task="SEND_EMAIL"
	);

/* 
Finally, we need to update our email flows to point to the landing page
flows.
*/ 

UPDATE flows 
SET 
	next_id = (
		SELECT f2.id
		FROM flows AS f2
		WHERE f2.previous_id=flows.id
	       		AND f2.campaign_id = flows.campaign_id
			AND f2.task = "LANDING_PAGE"
		)
	WHERE task = "SEND_EMAIL"	

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

