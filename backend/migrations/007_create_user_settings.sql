-- +goose Up
CREATE TABLE IF NOT EXISTS user_settings (
    id CHAR(36) PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    latitude REAL,
    longitude REAL,
    calendar_ics_url TEXT,
    timezone TEXT,
    news_categories TEXT CHECK (news_categories IS NULL OR json_valid(news_categories)),
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    deleted_at DATETIME
);

-- +goose StatementBegin
CREATE TRIGGER user_settings_updated_at
AFTER UPDATE ON user_settings
FOR EACH ROW
BEGIN
    UPDATE user_settings SET updated_at = datetime('now') WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS user_settings_updated_at;
DROP TABLE IF EXISTS user_settings;
