-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS link_visits (
    id SERIAL PRIMARY KEY,
    link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip VARCHAR(45) NOT NULL,
    user_agent TEXT,
    referer TEXT,
    status INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_link_visits_link_id ON link_visits(link_id);
CREATE INDEX idx_link_visits_created_at ON link_visits(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_link_visits_created_at;
DROP INDEX IF EXISTS idx_link_visits_link_id;
DROP TABLE IF EXISTS link_visits;
-- +goose StatementEnd
