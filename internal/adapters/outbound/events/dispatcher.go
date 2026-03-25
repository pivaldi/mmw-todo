package events

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/pivaldi/mmw-todo/internal/domain"
)

// LogEventDispatcher is a simple event dispatcher that logs events
type LogEventDispatcher struct {
	logger *slog.Logger
}

// NewLogEventDispatcher creates a new InMemoryEventDispatcher
func NewLogEventDispatcher(logger *slog.Logger) *LogEventDispatcher {
	return &LogEventDispatcher{
		logger: logger,
	}
}

// Dispatch publishes domain events
func (d *LogEventDispatcher) Dispatch(_ context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		eventData, err := json.Marshal(event)
		if err != nil {
			d.logger.Error("failed to marshal event",
				"error", err,
				"event_type", event.EventType(),
			)

			continue
		}

		d.logger.Info("domain event dispatched", "data", eventData)
	}

	return nil
}
