-- +goose Up
CREATE TABLE stocks_watchlist (
    id         CHAR(36) NOT NULL PRIMARY KEY CHECK(length(id) = 36),  -- UUID v7, app-generated
    user_id    CHAR(36) REFERENCES users(id),                          -- NULL until auth is wired; FK enforced
    symbol     TEXT     NOT NULL CHECK(length(trim(symbol)) > 0),
    created_at TEXT     NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT     NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    deleted_at TEXT                                                     -- NULL = active, soft-delete
);

-- Full (non-partial) unique index so that there is at most one row per (user_id, symbol)
-- across all time.  This lets the application re-activate a previously deleted row via UPSERT:
--
--   INSERT INTO stocks_watchlist (id, user_id, symbol)
--   VALUES (?, ?, ?)
--   ON CONFLICT(user_id, symbol) DO UPDATE
--     SET deleted_at = NULL,
--         updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now');
--
-- Without a full unique index a partial one (WHERE deleted_at IS NULL) would allow new rows on
-- every re-add cycle, bloating the table indefinitely.
--
-- NOTE: SQLite treats NULL as distinct in unique indexes, so (NULL, 'AAPL') does not conflict
-- with another (NULL, 'AAPL').  The UPSERT de-duplication only works correctly once auth is
-- wired and user_id is always non-NULL.
CREATE UNIQUE INDEX uq_stocks_watchlist_user_symbol
    ON stocks_watchlist (user_id, symbol);

-- +goose StatementBegin
CREATE TRIGGER stocks_watchlist_updated_at
    AFTER UPDATE ON stocks_watchlist
    FOR EACH ROW
    WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE stocks_watchlist
       SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
     WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
-- Note: roll back any migrations that FK-reference stocks_watchlist before this.
DROP TRIGGER IF EXISTS stocks_watchlist_updated_at;
DROP INDEX  IF EXISTS uq_stocks_watchlist_user_symbol;
DROP TABLE  IF EXISTS stocks_watchlist;
