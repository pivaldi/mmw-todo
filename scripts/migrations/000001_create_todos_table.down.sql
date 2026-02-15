-- Drop todos table and all associated indexes
DROP INDEX IF EXISTS idx_todos_priority;
DROP INDEX IF EXISTS idx_todos_created_at;
DROP INDEX IF EXISTS idx_todos_due_date;
DROP INDEX IF EXISTS idx_todos_status;
DROP TABLE IF EXISTS todos;
