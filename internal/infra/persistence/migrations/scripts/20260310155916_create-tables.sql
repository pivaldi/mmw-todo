-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS todos (
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
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_todos_created_at ON todos(created_at DESC);
CREATE INDEX idx_todos_priority ON todos(priority);

-- Add a comment to the table
COMMENT ON TABLE todos IS 'Stores todo items with their properties and status';

-- Create outbox events table for transactional event publishing via workers
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE
);

-- Index for finding unpublished events (used by outbox relay worker)
CREATE INDEX idx_unpublished ON events(occurred_at ASC) WHERE published_at IS NULL;

-- Index for published events cleanup
CREATE INDEX idx_published ON events(published_at) WHERE published_at IS NOT NULL;

-- Add comments
COMMENT ON TABLE events IS 'Transactional outbox for event publishing via workers';
COMMENT ON COLUMN events.payload IS 'Event payload';
COMMENT ON COLUMN events.published_at IS 'NULL indicates unpublished event, non-NULL means published';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_todos_priority;
DROP INDEX IF EXISTS idx_todos_created_at;
DROP INDEX IF EXISTS idx_todos_due_date;
DROP INDEX IF EXISTS idx_todos_status;
DROP TABLE IF EXISTS todos;

DROP INDEX IF EXISTS idx_published;
DROP INDEX IF EXISTS idx_unpublished;
DROP TABLE IF EXISTS events;

-- +goose StatementEnd
