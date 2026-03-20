-- +goose Up
-- +goose StatementBegin

-- No FK constraint: todo.todo and auth.users live in separate schemas managed
-- by separate services. A cross-schema FK would couple migrations.
--
-- Existing rows are development/seed data only. They are intentionally orphaned
-- by the sentinel UUID and will be inaccessible to any real user.
ALTER TABLE todo.todo ADD COLUMN user_id UUID;
UPDATE todo.todo SET user_id = '00000000-0000-0000-0000-000000000000' WHERE user_id IS NULL;
ALTER TABLE todo.todo ALTER COLUMN user_id SET NOT NULL;
CREATE INDEX idx_todo_user_id ON todo.todo(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_todo_user_id;
ALTER TABLE todo.todo DROP COLUMN user_id;

-- +goose StatementEnd
