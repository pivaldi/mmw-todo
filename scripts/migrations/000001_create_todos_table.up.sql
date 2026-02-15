-- Create todos table with constraints
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
