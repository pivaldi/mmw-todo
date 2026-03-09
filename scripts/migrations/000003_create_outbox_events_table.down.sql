-- Drop outbox events table and indexes
DROP INDEX IF EXISTS idx_outbox_published;
DROP INDEX IF EXISTS idx_outbox_unpublished;
DROP TABLE IF EXISTS outbox_events;
