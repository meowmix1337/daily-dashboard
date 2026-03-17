-- +goose Up
CREATE TABLE IF NOT EXISTS task_labels (
    id CHAR(36) PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#6366f1',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    deleted_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_task_labels_user_name
    ON task_labels(user_id, lower(name)) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE TRIGGER task_labels_updated_at
AFTER UPDATE ON task_labels
FOR EACH ROW
BEGIN
    UPDATE task_labels SET updated_at = datetime('now') WHERE id = NEW.id;
END;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS task_label_assignments (
    id CHAR(36) PRIMARY KEY,
    task_id CHAR(36) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id CHAR(36) NOT NULL REFERENCES task_labels(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    deleted_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_task_label_assignments_task_label
    ON task_label_assignments(task_id, label_id) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE TRIGGER task_label_assignments_updated_at
AFTER UPDATE ON task_label_assignments
FOR EACH ROW
BEGIN
    UPDATE task_label_assignments SET updated_at = datetime('now') WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS task_label_assignments_updated_at;
DROP TABLE IF EXISTS task_label_assignments;
DROP TRIGGER IF EXISTS task_labels_updated_at;
DROP TABLE IF EXISTS task_labels;
