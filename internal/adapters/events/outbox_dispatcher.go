// services/todo/internal/adapters/events/outbox_dispatcher.go
package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ovya/ogl/postgres/uow"
	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

type PostgresOutboxDispatcher struct {
	pool *pgxpool.Pool
}

func NewPostgresOutboxDispatcher(pool *pgxpool.Pool) *PostgresOutboxDispatcher {
	return &PostgresOutboxDispatcher{pool: pool}
}

// Dispatch saves all events to the outbox table efficiently using a single batch
func (d *PostgresOutboxDispatcher) Dispatch(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `INSERT INTO outbox_events (aggregate_id, event_type, payload, occurred_at) VALUES ($1, $2, $3, $4)`

	// Queue all events into the batch
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %s: %w", event.EventType(), err)
		}

		batch.Queue(query, event.AggregateID(), event.EventType(), payload, event.OccurredAt())
	}

	// Magically uses the transaction from the UoW context!
	exec := uow.GetExecutor(ctx, d.pool)

	br := exec.SendBatch(ctx, batch)
	defer br.Close()

	// Verify all inserts succeeded
	for i := range events {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to insert outbox event at index %d: %w", i, err)
		}
	}

	return nil
}
