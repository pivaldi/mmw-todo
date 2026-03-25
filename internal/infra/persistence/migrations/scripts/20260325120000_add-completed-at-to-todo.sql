-- +goose Up
ALTER TABLE todo.todo ADD COLUMN completed_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE todo.todo DROP COLUMN completed_at;
