-- +goose Up
CREATE TABLE task_priorities (
    id         TEXT NOT NULL PRIMARY KEY,   -- 'low', 'medium', 'high'
    label      TEXT NOT NULL,
    sort_order INTEGER NOT NULL             -- lower number = higher priority (1=high, 2=medium, 3=low)
);

INSERT INTO task_priorities (id, label, sort_order) VALUES
    ('high',   'High',   1),
    ('medium', 'Medium', 2),
    ('low',    'Low',    3);

-- +goose Down
DROP TABLE IF EXISTS task_priorities;
