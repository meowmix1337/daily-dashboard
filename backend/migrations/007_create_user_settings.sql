-- +goose Up
CREATE TABLE IF NOT EXISTS user_settings (
    id CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),
    user_id CHAR(36) NOT NULL REFERENCES users(id),
    latitude REAL CHECK (latitude IS NULL OR (latitude >= -90.0 AND latitude <= 90.0)),
    longitude REAL CHECK (longitude IS NULL OR (longitude >= -180.0 AND longitude <= 180.0)),
    -- NOTE: calendar_ics_url contains auth tokens (e.g. Google/Outlook private feed URLs).
    -- This column MUST be encrypted at the application layer before storage.
    -- See bd task: implement EncryptionService with Encrypt/Decrypt for this field.
    calendar_ics_url TEXT,
    timezone TEXT,
    news_categories TEXT CHECK (news_categories IS NULL OR json_valid(news_categories)),
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

-- +goose Down
DROP TRIGGER IF EXISTS user_settings_updated_at;
DROP INDEX  IF EXISTS uq_user_settings_user_id_active;
DROP TABLE  IF EXISTS user_settings;
