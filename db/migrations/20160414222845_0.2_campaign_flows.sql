
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE campaigns RENAME TO campaigns_old;
/*
The new campaigns table will rely on a "tasks" table that is basically a list of tasks to accomplish.

This will include sending emails, hosting landing pages, or possibly taking other action as needed,
including running scripts, sending HTTP requests, etc.
*/
CREATE TABLE campaigns ("id" integer primary key autoincrement, "user_id" bigint, "name" varchar(255) NOT NULL, "created_date" datetime, "completed_date" datetime, "task_id" bigint, "status" varchar(255)); 

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

/* Create our tasks table */

CREATE TABLE "tasks" ("id" integer primary key autoincrement,"user_id" bigint,"previous_id" bigint,"next_id" bigint,"campaign_id" bigint, "metadata" blob,"type" varchar(255));

/* Setup our email tasks */

INSERT INTO tasks (
	user_id,
	campaign_id,
	type,
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

/* Point our campaigns to the SMTP tasks */
UPDATE campaigns
SET
	task_id	= (SELECT tasks.id FROM tasks
		WHERE tasks.campaign_id = campaigns.id
		AND tasks.type = "SEND_EMAIL"
	);

/* Setup our landing page tasks */

INSERT INTO tasks (
	user_id,
	campaign_id,
	type,
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
	tasks.id as task_id
FROM campaigns_old, tasks
WHERE tasks.id IN (
	SELECT id FROM tasks
	WHERE campaign_id=campaigns_old.id
		AND type="SEND_EMAIL"
	);

/* 
Next, we need to update our email tasks to point to the landing page
tasks.
*/ 

UPDATE tasks 
SET 
	next_id = (
		SELECT t2.id
		FROM tasks AS t2
		WHERE t2.previous_id=tasks.id
	       		AND t2.campaign_id = tasks.campaign_id
			AND t2.type = "LANDING_PAGE"
		)
	WHERE type = "SEND_EMAIL";	

/* Finally, we drop our temp table */
DROP TABLE campaigns_old;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

