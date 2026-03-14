-- +goose Up
CREATE TABLE users (
    id         CHAR(36) NOT NULL PRIMARY KEY,            -- UUID v7 generated in application
    google_id  TEXT     NOT NULL UNIQUE,                -- Google OAuth sub claim
    email      TEXT     NOT NULL,
    name       TEXT     NOT NULL,
    avatar_url TEXT     NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at DATETIME                                 -- NULL = active, non-NULL = soft-deleted
);

-- +goose Down
DROP TABLE users;
