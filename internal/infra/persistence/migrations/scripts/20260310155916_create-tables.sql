-- +goose Up
-- +goose StatementBegin

-- Create schema todo
CREATE SCHEMA todo;

-- Create table todo.todo
CREATE TABLE IF NOT EXISTS todo.todo (
    id UUID PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT valid_status CHECK (status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    CONSTRAINT future_due_date CHECK (due_date IS NULL OR due_date > created_at)
);

-- Create indexes for common queries
CREATE INDEX idx_todo_status ON todo.todo(status);
CREATE INDEX idx_todo_due_date ON todo.todo(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_todo_created_at ON todo.todo(created_at DESC);
CREATE INDEX idx_todo_priority ON todo.todo(priority);

-- Add a comment to the table
COMMENT ON TABLE todo.todo IS 'Stores todo items with their properties and status';

-- Create outbox events table for transactional event publishing via workers
CREATE TABLE IF NOT EXISTS todo.event (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE
);

-- Index for finding unpublished events (used by outbox relay worker)
CREATE INDEX idx_unpublished ON todo.event(occurred_at ASC) WHERE published_at IS NULL;

-- Index for published events cleanup
CREATE INDEX idx_published ON todo.event(published_at) WHERE published_at IS NOT NULL;

-- Add comments
COMMENT ON TABLE todo.event IS 'Transactional outbox for event publishing via workers';
COMMENT ON COLUMN todo.event.payload IS 'Event payload';
COMMENT ON COLUMN todo.event.published_at IS 'NULL indicates unpublished event, non-NULL means published';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_todo_priority;
DROP INDEX IF EXISTS idx_todo_created_at;
DROP INDEX IF EXISTS idx_todo_due_date;
DROP INDEX IF EXISTS idx_todo_status;
DROP TABLE IF EXISTS todo;

DROP INDEX IF EXISTS idx_published;
DROP INDEX IF EXISTS idx_unpublished;
DROP TABLE IF EXISTS todo.event;

-- +goose StatementEnd
