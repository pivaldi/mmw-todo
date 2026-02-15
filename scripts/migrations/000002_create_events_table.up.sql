-- Create domain events outbox table for transactional event publishing
CREATE TABLE IF NOT EXISTS domain_events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    published_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT valid_event_type CHECK (event_type IN (
        'TodoCreated',
        'TodoUpdated',
        'TodoCompleted',
        'TodoReopened',
        'TodoDeleted'
    ))
);

-- Index for finding unpublished events
CREATE INDEX idx_unpublished_events ON domain_events(published_at) WHERE published_at IS NULL;

-- Index for querying events by aggregate
CREATE INDEX idx_events_by_aggregate ON domain_events(aggregate_id, occurred_at DESC);

-- Index for querying by event type
CREATE INDEX idx_events_by_type ON domain_events(event_type, occurred_at DESC);

-- Add comments
COMMENT ON TABLE domain_events IS 'Transactional outbox for domain events';
COMMENT ON COLUMN domain_events.published_at IS 'NULL indicates unpublished event';
