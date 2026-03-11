package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotisserie/eris"
)

// SystemEventBus represents the actual transport mechanism across the monolith
// e.g., an in-memory channel broker, NATS, or RabbitMQ.
type SystemEventBus interface {
	Publish(ctx context.Context, eventType string, payload []byte) error
}

type EventsRelay struct {
	pool     *pgxpool.Pool
	bus      SystemEventBus
	logger   *slog.Logger
	interval time.Duration
}

func NewEnventsRelay(pool *pgxpool.Pool, bus SystemEventBus, logger *slog.Logger) *EventsRelay {
	return &EventsRelay{
		pool:     pool,
		bus:      bus,
		logger:   logger,
		interval: 2 * time.Second, // Poll every 2 seconds
	}
}

// Start runs continuously until the context is canceled (Graceful Shutdown)
func (r *EventsRelay) Start(ctx context.Context) {
	r.logger.Info("starting outbox relay worker")
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("shutting down outbox relay worker")
			return
		case <-ticker.C:
			if err := r.processBatch(ctx); err != nil {
				r.logger.Error("outbox processing failed", "error", err)
			}
		}
	}
}

func (r *EventsRelay) processBatch(ctx context.Context) error {
	// 1. Open a transaction for the worker
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return eris.Wrap(err, "opening worker transaction failed")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 2. Fetch unpublished events (Lock them so other workers ignore them)
	query := `
		SELECT id, event_type, payload
		FROM events
		WHERE published_at IS NULL
		ORDER BY occurred_at ASC
		LIMIT 100
		FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return eris.Wrap(err, "fetching unpublished events failed")
	}
	defer rows.Close()

	var eventIDs []int

	for rows.Next() {
		var id int
		var eventType string
		var payload []byte

		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			return eris.Wrap(err, "scaning unpublished events failed")
		}

		// 3. Publish to the real system bus
		if err := r.bus.Publish(ctx, eventType, payload); err != nil {
			// If publishing fails, we return. The defer tx.Rollback() unlocks the rows
			// so they can be retried on the next tick!
			return eris.Wrap(err, "publishing evants fails")
		}

		eventIDs = append(eventIDs, id)
	}
	rows.Close() // Must close rows before executing the next query

	// 4. If we published anything, mark them as done using pgx.Batch
	if len(eventIDs) > 0 {
		batch := &pgx.Batch{}
		updateQuery := `UPDATE events SET published_at = NOW() WHERE id = $1`
		for _, id := range eventIDs {
			batch.Queue(updateQuery, id)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		for i := 0; i < len(eventIDs); i++ {
			if _, err := br.Exec(); err != nil {
				return eris.Wrap(err, "mark published event as done failed")
			}
		}
		br.Close()

		r.logger.Info("processed outbox batch", "count", len(eventIDs))
	}

	// 5. Commit the transaction to finalize the published_at updates
	err = tx.Commit(ctx)
	if err != nil {
		return eris.Wrap(err, "commiting event worker failed")
	}

	return nil
}
