
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE "roles" (
    "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
    "slug"        VARCHAR(255) NOT NULL UNIQUE,
    "name"        VARCHAR(255) NOT NULL UNIQUE,
    "description" VARCHAR(255)
);

ALTER TABLE "users" ADD COLUMN "role_id" INTEGER;

CREATE TABLE "permissions" (
    "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
    "slug"        VARCHAR(255) NOT NULL UNIQUE,
    "name"        VARCHAR(255) NOT NULL UNIQUE,
    "description" VARCHAR(255)
);


CREATE TABLE "role_permissions" (
    "role_id"       INTEGER NOT NULL,
    "permission_id" INTEGER NOT NULL
);

INSERT INTO "roles" ("slug", "name", "description")
VALUES
    ("admin", "Admin", "System administrator with full permissions"),
    ("user", "User", "User role with edit access to objects and campaigns");

INSERT INTO "permissions" ("slug", "name", "description")
VALUES
    ("view_objects", "View Objects", "View objects in Gophish"),
    ("modify_objects", "Modify Objects", "Create and edit objects in Gophish"),
    ("modify_system", "Modify System", "Manage system-wide configuration");

-- Our rules for generating the admin user are:
-- * The user with the name "admin"
-- * OR the first user, if no "admin" user exists
UPDATE "users" SET "role_id"=(
    SELECT "id" FROM "roles" WHERE "slug"="admin")
WHERE "id"=(
    SELECT "id" FROM "users" WHERE "username"="admin" OR "id"=(SELECT MIN("id") FROM "users") LIMIT 1);

-- Every other user will be considered a standard user account. The admin user
-- will be able to change the role of any other user at any time.
UPDATE "users" SET "role_id"=(
    SELECT "id" FROM "roles" WHERE "slug"="user")
WHERE role_id IS NULL;

-- Our default permission set will:
-- * Allow admins the ability to do anything
-- * Allow users to modify objects

-- Allow any user to view objects
INSERT INTO "role_permissions" ("role_id", "permission_id")
SELECT r.id, p.id FROM roles AS r, "permissions" AS p
WHERE r.id IN (SELECT "id" FROM roles WHERE "slug"="admin" OR "slug"="user")
AND p.id=(SELECT "id" FROM "permissions" WHERE "slug"="view_objects");

-- Allow admins and users to modify objects
INSERT INTO "role_permissions" ("role_id", "permission_id")
SELECT r.id, p.id FROM roles AS r, "permissions" AS p
WHERE r.id IN (SELECT "id" FROM roles WHERE "slug"="admin" OR "slug"="user")
AND p.id=(SELECT "id" FROM "permissions" WHERE "slug"="modify_objects");

-- Allow admins to modify system level configuration
INSERT INTO "role_permissions" ("role_id", "permission_id")
SELECT r.id, p.id FROM roles AS r, "permissions" AS p
WHERE r.id IN (SELECT "id" FROM roles WHERE "slug"="admin")
AND p.id=(SELECT "id" FROM "permissions" WHERE "slug"="modify_system");

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "roles"
DROP TABLE "user_roles"
DROP TABLE "permissions"
