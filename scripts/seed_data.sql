-- Sample seed data for development and testing
-- This file can be run after migrations to populate the database with test data

-- Insert sample todos
INSERT INTO todos (id, title, description, status, priority, due_date, created_at, updated_at)
VALUES
    (
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        'Complete project documentation',
        'Write comprehensive documentation for the Todo API including architecture decisions and usage examples',
        'in_progress',
        'high',
        NOW() + INTERVAL '7 days',
        NOW() - INTERVAL '2 days',
        NOW() - INTERVAL '1 day'
    ),
    (
        'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12',
        'Review pull requests',
        'Review and merge pending pull requests from team members',
        'pending',
        'medium',
        NOW() + INTERVAL '3 days',
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '1 day'
    ),
    (
        'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13',
        'Fix authentication bug',
        'Investigate and fix the issue where users are logged out unexpectedly',
        'completed',
        'urgent',
        NOW() - INTERVAL '2 days',
        NOW() - INTERVAL '5 days',
        NOW() - INTERVAL '1 day'
    ),
    (
        'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a14',
        'Update dependencies',
        'Update all project dependencies to their latest stable versions',
        'pending',
        'low',
        NOW() + INTERVAL '14 days',
        NOW(),
        NOW()
    ),
    (
        'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15',
        'Prepare demo presentation',
        'Create slides and demo for the upcoming stakeholder meeting',
        'pending',
        'high',
        NOW() + INTERVAL '5 days',
        NOW(),
        NOW()
    );

-- Insert corresponding domain events
INSERT INTO domain_events (aggregate_id, event_type, event_data, occurred_at, published_at)
VALUES
    (
        'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        'TodoCreated',
        '{"id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", "title": "Complete project documentation", "priority": "high"}'::jsonb,
        NOW() - INTERVAL '2 days',
        NOW() - INTERVAL '2 days'
    ),
    (
        'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12',
        'TodoCreated',
        '{"id": "b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", "title": "Review pull requests", "priority": "medium"}'::jsonb,
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '1 day'
    ),
    (
        'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13',
        'TodoCreated',
        '{"id": "c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13", "title": "Fix authentication bug", "priority": "urgent"}'::jsonb,
        NOW() - INTERVAL '5 days',
        NOW() - INTERVAL '5 days'
    ),
    (
        'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13',
        'TodoCompleted',
        '{"id": "c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13", "completed_at": "' || (NOW() - INTERVAL '1 day')::text || '"}'::jsonb,
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '1 day'
    );
