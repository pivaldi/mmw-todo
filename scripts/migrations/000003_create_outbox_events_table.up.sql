-- Create outbox events table for transactional event publishing via workers
CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    payload BYTEA NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE
);

-- Index for finding unpublished events (used by outbox relay worker)
CREATE INDEX idx_outbox_unpublished ON outbox_events(occurred_at ASC) WHERE published_at IS NULL;

-- Index for published events cleanup
CREATE INDEX idx_outbox_published ON outbox_events(published_at) WHERE published_at IS NOT NULL;

-- Add comments
COMMENT ON TABLE outbox_events IS 'Transactional outbox for event publishing via workers';
COMMENT ON COLUMN outbox_events.payload IS 'Raw event payload in bytes';
COMMENT ON COLUMN outbox_events.published_at IS 'NULL indicates unpublished event, non-NULL means published';
