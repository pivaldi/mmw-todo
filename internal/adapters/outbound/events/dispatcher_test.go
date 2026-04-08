package events

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

func TestInMemoryEventDispatcher_Dispatch_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := NewLogEventDispatcher(logger)

	// Create test events
	todoID := domain.NewTodoID()
	userID := uuid.New()
	title, _ := domain.NewTaskTitle("Test Todo")
	event := domain.NewTodoCreatedEvent(todoID, userID, title, "Description", domain.PriorityMedium, nil)

	events := []domain.DomainEvent{event}

	// Should not return error
	err := dispatcher.Dispatch(context.Background(), events)
	if err != nil {
		t.Errorf("Dispatch() unexpected error: %v", err)
	}
}

func TestInMemoryEventDispatcher_Dispatch_EmptyEvents_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := NewLogEventDispatcher(logger)

	// Should handle empty event slice
	err := dispatcher.Dispatch(context.Background(), []domain.DomainEvent{})
	if err != nil {
		t.Errorf("Dispatch() unexpected error for empty events: %v", err)
	}
}
