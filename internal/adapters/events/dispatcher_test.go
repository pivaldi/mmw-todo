package events

import (
	"context"
	"log/slog"
	"os"
	"testing"

	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
)

func TestInMemoryEventDispatcher_Dispatch_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := NewInMemoryEventDispatcher(logger)

	// Create test events
	todoID := domain.NewTodoID()
	title, _ := domain.NewTaskTitle("Test Todo")
	event := domain.NewTodoCreatedEvent(todoID, title, "Description", domain.PriorityMedium, nil)

	events := []domain.DomainEvent{event}

	// Should not return error
	err := dispatcher.Dispatch(context.Background(), events)

	if err != nil {
		t.Errorf("Dispatch() unexpected error: %v", err)
	}
}

func TestInMemoryEventDispatcher_Dispatch_EmptyEvents_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := NewInMemoryEventDispatcher(logger)

	// Should handle empty event slice
	err := dispatcher.Dispatch(context.Background(), []domain.DomainEvent{})

	if err != nil {
		t.Errorf("Dispatch() unexpected error for empty events: %v", err)
	}
}

func TestInMemoryEventDispatcher_Dispatch_MultipleEvents_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dispatcher := NewInMemoryEventDispatcher(logger)

	// Create multiple test events
	todoID := domain.NewTodoID()
	title, _ := domain.NewTaskTitle("Test Todo")
	event1 := domain.NewTodoCreatedEvent(todoID, title, "Description", domain.PriorityMedium, nil)
	event2 := domain.NewTodoUpdatedEvent(todoID)

	events := []domain.DomainEvent{event1, event2}

	// Should handle multiple events
	err := dispatcher.Dispatch(context.Background(), events)

	if err != nil {
		t.Errorf("Dispatch() unexpected error for multiple events: %v", err)
	}
}
