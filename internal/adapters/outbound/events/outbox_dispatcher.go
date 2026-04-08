package events

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	pfuow "github.com/piprim/mmw/pkg/platform/pg/uow"
	todov1 "github.com/pivaldi/mmw-contracts/gen/go/todo/v1"
	"github.com/pivaldi/mmw-todo/internal/domain"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostgresOutboxDispatcher struct {
	uow *pfuow.UnitOfWork
}

func NewPostgresOutboxDispatcher(uow *pfuow.UnitOfWork) *PostgresOutboxDispatcher {
	return &PostgresOutboxDispatcher{uow: uow}
}

// Dispatch saves all events to the outbox table efficiently using a single batch.
// The stored event_type is the Watermill routing key, resolved via domainTopics.
func (d *PostgresOutboxDispatcher) Dispatch(ctx context.Context, events []domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `INSERT INTO todo.event (event_type, payload, occurred_at) VALUES ($1, $2::jsonb, $3)`

	// Queue all events into the batch
	for _, event := range events {
		topic, ok := domainTopics[event.EventType()]
		if !ok {
			return eris.Errorf("no routing key for domain event type %q", event.EventType())
		}

		var payload []byte
		var err error

		// Map domain events to proto messages where possible
		if protoMsg := domainToProto(event); protoMsg != nil {
			payload, err = protojson.Marshal(protoMsg)
		} else {
			payload, err = json.Marshal(event)
		}

		if err != nil {
			return eris.Wrapf(err, "failed to marshal event %s", event.EventType())
		}

		batch.Queue(query, topic, string(payload), event.GetOccurredAt())
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

func domainToProto(event domain.DomainEvent) protoreflect.ProtoMessage {
	switch e := event.(type) {
	case *domain.UserTasksDeleted:
		return &todov1.UserTasksDeletedEvent{
			UserId:    e.GetAggregateID(),
			DeletedAt: timestamppb.New(e.GetOccurredAt()),
		}
	case *domain.TodoCreated:
		return &todov1.UserTaskCreatedEvent{
			UserId:    e.GetUserID().String(),
			TaskId:    e.GetAggregateID(),
			CreatedAt: timestamppb.New(e.GetOccurredAt()),
		}
	case *domain.TodoUpdated:
		return &todov1.UserTaskUpdatedEvent{
			UserId:    e.GetUserID().String(),
			TaskId:    e.GetAggregateID(),
			UpdatedAt: timestamppb.New(e.GetOccurredAt()),
		}
	case *domain.TodoCompleted:
		return &todov1.UserTaskCompletedEvent{
			UserId:      e.GetUserID().String(),
			TaskId:      e.GetAggregateID(),
			CompletedAt: timestamppb.New(e.CompletedAt),
		}
	case *domain.TodoReopened:
		return &todov1.UserTaskReopenedEvent{
			UserId:     e.GetUserID().String(),
			TaskId:     e.GetAggregateID(),
			ReopenedAt: timestamppb.New(e.GetOccurredAt()),
		}
	case *domain.TodoDeleted:
		return &todov1.UserTaskDeletedEvent{
			UserId:    e.GetUserID().String(),
			TaskId:    e.GetAggregateID(),
			DeletedAt: timestamppb.New(e.GetOccurredAt()),
		}
	}

	return nil
}
