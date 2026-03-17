-- +goose Up
CREATE TABLE IF NOT EXISTS task_labels (
    id CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),
    user_id CHAR(36) NOT NULL REFERENCES users(id),
    name TEXT NOT NULL CHECK(length(trim(name)) > 0),
    color TEXT NOT NULL DEFAULT '#6366f1',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_task_labels_user_id
    ON task_labels(user_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_task_labels_user_name
    ON task_labels(user_id, lower(name)) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE TRIGGER task_labels_updated_at
    AFTER UPDATE ON task_labels
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE task_labels
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = NEW.id;
END;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS task_label_assignments (
    id CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),
    task_id CHAR(36) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id CHAR(36) NOT NULL REFERENCES task_labels(id) ON DELETE CASCADE,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_task_label_assignments_task_label
    ON task_label_assignments(task_id, label_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_task_label_assignments_label_id
    ON task_label_assignments(label_id) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE TRIGGER task_label_assignments_updated_at
    AFTER UPDATE ON task_label_assignments
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE task_label_assignments
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS task_label_assignments_updated_at;
DROP INDEX  IF EXISTS idx_task_label_assignments_label_id;
DROP INDEX  IF EXISTS idx_task_label_assignments_task_label;
DROP TABLE  IF EXISTS task_label_assignments;
DROP TRIGGER IF EXISTS task_labels_updated_at;
DROP INDEX  IF EXISTS idx_task_labels_user_name;
DROP INDEX  IF EXISTS idx_task_labels_user_id;
DROP TABLE  IF EXISTS task_labels;
