-- +goose Up
CREATE TABLE tasks (
    id          CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),  -- UUID v7, app-generated
    user_id     CHAR(36) REFERENCES users(id),                          -- NULL until auth is wired; FK enforced
    text        TEXT     NOT NULL CHECK(length(trim(text)) > 0),
    done        INTEGER  NOT NULL DEFAULT 0 CHECK(done IN (0, 1)),      -- SQLite boolean
    priority_id TEXT     NOT NULL DEFAULT 'medium' REFERENCES task_priorities(id),
    created_at  TEXT     NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT     NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at  TEXT                                                     -- NULL = active, soft-delete
);

CREATE INDEX idx_tasks_user_id ON tasks (user_id) WHERE deleted_at IS NULL;

-- +goose StatementBegin
CREATE TRIGGER tasks_updated_at
    AFTER UPDATE ON tasks
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE tasks
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
-- Note: roll back any migrations that FK-reference tasks before this.
DROP TRIGGER IF EXISTS tasks_updated_at;
DROP INDEX  IF EXISTS idx_tasks_user_id;
DROP TABLE  IF EXISTS tasks;
