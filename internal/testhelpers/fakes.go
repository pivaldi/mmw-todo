// Package testhelpers provides in-memory fakes for ports.TodoRepository,
// ports.UnitOfWork, and ports.EventDispatcher. Use these in unit and contract
// tests to wire a real TodoApplicationService without any infrastructure.
package testhelpers

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// --- InMemoryTodoRepo ---

// InMemoryTodoRepo is a map-backed implementation of ports.TodoRepository.
// It returns domain.ErrTodoNotFound when an item is absent, and scopes
// FindByID and Delete to the provided userID.
type InMemoryTodoRepo struct {
	mu    sync.RWMutex
	todos map[domain.TodoID]*domain.Todo
}

func NewInMemoryTodoRepo() *InMemoryTodoRepo {
	return &InMemoryTodoRepo{
		todos: make(map[domain.TodoID]*domain.Todo),
	}
}

func (r *InMemoryTodoRepo) Save(_ context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.todos[todo.ID()] = todo
	return nil
}

func (r *InMemoryTodoRepo) FindByID(_ context.Context, id domain.TodoID, userID uuid.UUID) (*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.todos[id]
	if !ok || t.UserID() != userID {
		return nil, domain.ErrTodoNotFound
	}
	return t, nil
}

func (r *InMemoryTodoRepo) FindAll(_ context.Context, filters ports.Filters) ([]*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Todo
	for _, t := range r.todos {
		if filters.UserID != nil && t.UserID() != *filters.UserID {
			continue
		}
		if filters.Status != nil && t.Status() != *filters.Status {
			continue
		}
		if filters.Priority != nil && t.Priority() != *filters.Priority {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (r *InMemoryTodoRepo) Update(_ context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.todos[todo.ID()]
	if !ok || t.UserID() != todo.UserID() {
		return domain.ErrTodoNotFound
	}
	r.todos[todo.ID()] = todo
	return nil
}

func (r *InMemoryTodoRepo) Delete(_ context.Context, id domain.TodoID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.todos[id]
	if !ok || t.UserID() != userID {
		return domain.ErrTodoNotFound
	}
	delete(r.todos, id)
	return nil
}

func (r *InMemoryTodoRepo) Health(_ context.Context) (any, error) {
	return 0, nil
}

// --- PassthroughUoW ---

// PassthroughUoW implements ports.UnitOfWork by calling fn(ctx) directly.
// There is no real transaction — it is suitable only for in-memory tests.
type PassthroughUoW struct{}

func (PassthroughUoW) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// --- NoopEventDispatcher ---

// NoopEventDispatcher implements ports.EventDispatcher by silently dropping all events.
type NoopEventDispatcher struct{}

func (NoopEventDispatcher) Dispatch(_ context.Context, _ []domain.DomainEvent) error {
	return nil
}

// --- Factory ---

// NewTestService creates a real TodoApplicationService wired with in-memory fakes.
// It returns both the service and the repo so tests can pre-seed state.
func NewTestService(t *testing.T) (application.TodoService, *InMemoryTodoRepo) {
	t.Helper()
	repo := NewInMemoryTodoRepo()
	svc := application.NewTodoApplicationService(repo, PassthroughUoW{}, NoopEventDispatcher{})
	return svc, repo
}
