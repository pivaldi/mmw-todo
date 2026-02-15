-- Drop domain events table and all associated indexes
DROP INDEX IF EXISTS idx_events_by_type;
DROP INDEX IF EXISTS idx_events_by_aggregate;
DROP INDEX IF EXISTS idx_unpublished_events;
DROP TABLE IF EXISTS domain_events;
