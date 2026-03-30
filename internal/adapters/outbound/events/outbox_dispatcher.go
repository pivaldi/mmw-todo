// services/todo/internal/adapters/events/outbox_dispatcher.go
package events

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	pfuow "github.com/piprim/mmw/platform/pg/uow"
	"github.com/pivaldi/mmw-todo/internal/domain"
	"github.com/rotisserie/eris"
)

type PostgresOutboxDispatcher struct {
	uow *pfuow.UnitOfWork
}

func NewPostgresOutboxDispatcher(uow *pfuow.UnitOfWork) *PostgresOutboxDispatcher {
	return &PostgresOutboxDispatcher{uow: uow}
}

// Dispatch saves all events to the outbox table efficiently using a single batch
func (d *PostgresOutboxDispatcher) Dispatch(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `INSERT INTO todo.event (event_type, payload, occurred_at) VALUES ($1, $2::jsonb, $3)`

	// Queue all events into the batch
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return eris.Wrapf(err, "failed to marshal event %s", event.EventType())
		}

		batch.Queue(query, event.EventType(), string(payload), event.GetOccurredAt())
	}

	exec := d.uow.Executor(ctx)

	br := exec.SendBatch(ctx, batch)
	defer br.Close()

	// Verify all inserts succeeded
	for i := range events {
		if _, err := br.Exec(); err != nil {
			return eris.Wrapf(err, "failed to insert outbox event at index %d", i)
		}
	}

	return nil
}
