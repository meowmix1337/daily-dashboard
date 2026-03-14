-- +goose Up
CREATE TABLE users (
    id         CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),       -- UUID v7, app-generated
    google_id  TEXT,                                                         -- NULL if not OAuth via Google
    email      TEXT     NOT NULL CHECK(email = lower(email) AND length(trim(email)) > 0),
    name       TEXT     NOT NULL CHECK(length(trim(name)) > 0),
    avatar_url TEXT              CHECK(avatar_url IS NULL OR length(avatar_url) > 0),
    created_at TEXT     NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT     NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at TEXT,                                                         -- NULL = active, soft-delete

    -- Ensure every user has at least one credential anchor.
    -- Extend to: OR password_hash IS NOT NULL when username/password auth is added.
    CHECK(google_id IS NOT NULL)
);

-- Partial unique indexes: enforce uniqueness only among active (non-deleted) rows.
-- Callers MUST include `AND deleted_at IS NULL` in lookups to engage these indexes.
-- Uses lower(email) to prevent case-variant duplicate emails.
CREATE UNIQUE INDEX uq_users_google_id_active ON users (google_id)      WHERE deleted_at IS NULL AND google_id IS NOT NULL;
CREATE UNIQUE INDEX uq_users_email_active      ON users (lower(email))  WHERE deleted_at IS NULL;

-- Auto-maintain updated_at on every UPDATE.
-- WHEN guard prevents re-firing when only updated_at itself changed (avoids infinite recursion).
-- +goose StatementBegin
CREATE TRIGGER users_updated_at
    AFTER UPDATE ON users
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE users
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
-- Note: any tables with FKs referencing users must be rolled back before this migration.
DROP TRIGGER IF EXISTS users_updated_at;
DROP INDEX  IF EXISTS uq_users_email_active;
DROP INDEX  IF EXISTS uq_users_google_id_active;
DROP TABLE  IF EXISTS users;
