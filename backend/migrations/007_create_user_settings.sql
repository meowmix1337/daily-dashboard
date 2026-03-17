-- +goose Up
CREATE TABLE IF NOT EXISTS user_settings (
    id CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),
    user_id CHAR(36) NOT NULL REFERENCES users(id),
    latitude REAL CHECK (latitude IS NULL OR (latitude >= -90.0 AND latitude <= 90.0)),
    longitude REAL CHECK (longitude IS NULL OR (longitude >= -180.0 AND longitude <= 180.0)),
    -- NOTE: calendar_ics_url contains auth tokens (e.g. Google/Outlook private feed URLs).
    -- This column MUST be encrypted at the application layer before storage.
    -- See bd task bd-75c: implement EncryptionService with Encrypt/Decrypt for this field.
    calendar_ics_url TEXT,
    timezone TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_settings_user_id_active
    ON user_settings (user_id) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE TRIGGER user_settings_updated_at
    AFTER UPDATE ON user_settings
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE user_settings
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- News category lookup table (GNews supported categories)
CREATE TABLE IF NOT EXISTS news_category_types (
    id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

INSERT INTO news_category_types (id, label, sort_order) VALUES
    ('general',       'General',       1),
    ('world',         'World',         2),
    ('nation',        'Nation',        3),
    ('business',      'Business',      4),
    ('technology',    'Technology',    5),
    ('entertainment', 'Entertainment', 6),
    ('sports',        'Sports',        7),
    ('science',       'Science',       8),
    ('health',        'Health',        9);

-- User selected news categories (M:N)
CREATE TABLE IF NOT EXISTS user_news_categories (
    id CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),
    user_id CHAR(36) NOT NULL REFERENCES users(id),
    category_id TEXT NOT NULL REFERENCES news_category_types(id),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_news_categories_user_category
    ON user_news_categories (user_id, category_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_news_categories_user_id
    ON user_news_categories (user_id) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE TRIGGER user_news_categories_updated_at
    AFTER UPDATE ON user_news_categories
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE user_news_categories
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS user_news_categories_updated_at;
DROP INDEX  IF EXISTS idx_user_news_categories_user_id;
DROP INDEX  IF EXISTS uq_user_news_categories_user_category;
DROP TABLE  IF EXISTS user_news_categories;
DROP TABLE  IF EXISTS news_category_types;
DROP TRIGGER IF EXISTS user_settings_updated_at;
DROP INDEX  IF EXISTS uq_user_settings_user_id_active;
DROP TABLE  IF EXISTS user_settings;
