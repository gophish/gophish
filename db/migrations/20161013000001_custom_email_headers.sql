
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS custom_headers(
	id integer primary key autoincrement,
	key varchar(255),
	value varchar(255),
	"template_id" bigint
);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE custom_headers;
