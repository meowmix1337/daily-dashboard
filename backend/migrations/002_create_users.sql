-- +goose Up
CREATE TABLE users (
    id         CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),  -- UUID v7, app-generated
    google_id  TEXT,                                                    -- NULL if not OAuth via Google
    email      TEXT     NOT NULL,
    name       TEXT     NOT NULL CHECK(length(trim(name)) > 0),
    avatar_url TEXT,                                                    -- NULL if no avatar provided
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at DATETIME                                                 -- NULL = active, soft-delete
);

-- Partial unique indexes: enforce uniqueness only among active (non-deleted) rows.
-- Allows re-registration after soft-delete without UNIQUE violations.
CREATE UNIQUE INDEX uq_users_google_id_active ON users (google_id) WHERE deleted_at IS NULL AND google_id IS NOT NULL;
CREATE UNIQUE INDEX uq_users_email_active      ON users (email)     WHERE deleted_at IS NULL;

-- Keep updated_at current on every UPDATE without requiring callers to remember.
-- +goose StatementBegin
CREATE TRIGGER users_updated_at
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS users_updated_at;
DROP INDEX  IF EXISTS uq_users_email_active;
DROP INDEX  IF EXISTS uq_users_google_id_active;
DROP TABLE  IF EXISTS users;
