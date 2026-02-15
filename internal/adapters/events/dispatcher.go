package events

import (
	"context"
	"encoding/json"
	"log/slog"

	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

// InMemoryEventDispatcher is a simple event dispatcher that logs events
// In production, this would publish to a message broker (RabbitMQ, Kafka, etc.)
type InMemoryEventDispatcher struct {
	logger *slog.Logger
}

// NewInMemoryEventDispatcher creates a new InMemoryEventDispatcher
func NewInMemoryEventDispatcher(logger *slog.Logger) *InMemoryEventDispatcher {
	return &InMemoryEventDispatcher{
		logger: logger,
	}
}

// Dispatch publishes domain events
// Currently logs events; in production would publish to message broker
func (d *InMemoryEventDispatcher) Dispatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		// Serialize event data for logging
		eventData, err := json.Marshal(map[string]interface{}{
			"type":         event.EventType(),
			"aggregate_id": event.AggregateID(),
			"occurred_at":  event.OccurredAt(),
		})
		if err != nil {
			d.logger.Error("failed to marshal event",
				"error", err,
				"event_type", event.EventType(),
			)
			continue
		}

		// Log the event
		d.logger.Info("domain event dispatched",
			"event_type", event.EventType(),
			"aggregate_id", event.AggregateID(),
			"occurred_at", event.OccurredAt(),
			"event_data", string(eventData),
		)
	}

	return nil
}
