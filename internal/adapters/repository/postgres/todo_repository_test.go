//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	domain "github.com/pivaldi/mmw/todo/internal/domain/todo"
	"github.com/pivaldi/mmw/todo/internal/ports"
)

var testDB *pgxpool.Pool

// setupTestDB creates a PostgreSQL container and runs migrations
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Clean up container when test completes
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %v", err)
		}
	})

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Clean up pool when test completes
	t.Cleanup(func() {
		pool.Close()
	})

	// Run migrations
	if err := runMigrations(ctx, pool); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return pool
}

// runMigrations executes migration files
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Get migrations directory
	migrationsDir := filepath.Join("..", "..", "..", "..", "scripts", "migrations")

	// Read up migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	// Execute .up.sql files in order
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".sql" && filepath.Ext(entry.Name()[:len(entry.Name())-4]) == ".up" {
			migrationPath := filepath.Join(migrationsDir, entry.Name())
			content, err := os.ReadFile(migrationPath)
			if err != nil {
				return fmt.Errorf("reading migration %s: %w", entry.Name(), err)
			}

			if _, err := pool.Exec(ctx, string(content)); err != nil {
				return fmt.Errorf("executing migration %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// createTestTodo creates a test todo for use in tests
func createTestTodo() *domain.Todo {
	title, _ := domain.NewTaskTitle("Test Todo")
	return domain.NewTodo(title, "Test description", domain.PriorityMedium, nil)
}

// createTestTodoWithDueDate creates a test todo with a due date
func createTestTodoWithDueDate() *domain.Todo {
	title, _ := domain.NewTaskTitle("Test Todo with Due Date")
	futureDate := time.Now().Add(24 * time.Hour)
	dueDate, _ := domain.NewDueDate(futureDate)
	return domain.NewTodo(title, "Test description", domain.PriorityHigh, &dueDate)
}

func TestPostgresTodRepository_Save_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	todo := createTestTodo()
	err := repo.Save(context.Background(), todo)

	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	// Verify todo was saved
	saved, err := repo.FindByID(context.Background(), todo.ID())
	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}

	if saved.ID() != todo.ID() {
		t.Errorf("ID = %v, want %v", saved.ID(), todo.ID())
	}

	if saved.Title().String() != todo.Title().String() {
		t.Errorf("Title = %v, want %v", saved.Title().String(), todo.Title().String())
	}
}

func TestPostgresTodRepository_Save_WithDueDate_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	todo := createTestTodoWithDueDate()
	err := repo.Save(context.Background(), todo)

	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	// Verify todo was saved with due date
	saved, err := repo.FindByID(context.Background(), todo.ID())
	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}

	if saved.DueDate() == nil {
		t.Fatal("Expected due date to be set")
	}

	// Compare timestamps (allow small difference due to precision)
	if todo.DueDate() != nil && saved.DueDate() != nil {
		diff := todo.DueDate().Time().Sub(saved.DueDate().Time())
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("DueDate difference too large: %v", diff)
		}
	}
}

func TestPostgresTodRepository_FindByID_NotFound_ReturnsError(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	nonExistentID := domain.NewTodoID()
	_, err := repo.FindByID(context.Background(), nonExistentID)

	if err == nil {
		t.Error("FindByID() expected error for non-existent todo, got nil")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("FindByID() error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestPostgresTodRepository_FindAll_NoFilters_ReturnsAll(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Create and save multiple todos
	todo1 := createTestTodo()
	todo2 := createTestTodo()
	todo3 := createTestTodo()

	if err := repo.Save(context.Background(), todo1); err != nil {
		t.Fatalf("Save() todo1 failed: %v", err)
	}
	if err := repo.Save(context.Background(), todo2); err != nil {
		t.Fatalf("Save() todo2 failed: %v", err)
	}
	if err := repo.Save(context.Background(), todo3); err != nil {
		t.Fatalf("Save() todo3 failed: %v", err)
	}

	// Find all
	todos, err := repo.FindAll(context.Background(), ports.Filters{})

	if err != nil {
		t.Fatalf("FindAll() unexpected error: %v", err)
	}

	if len(todos) != 3 {
		t.Errorf("FindAll() returned %d todos, want 3", len(todos))
	}
}

func TestPostgresTodRepository_FindAll_WithStatusFilter_FiltersCorrectly(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Create todos with different statuses
	todo1 := createTestTodo()
	todo2 := createTestTodo()
	todo2.Complete()

	if err := repo.Save(context.Background(), todo1); err != nil {
		t.Fatalf("Save() todo1 failed: %v", err)
	}
	if err := repo.Save(context.Background(), todo2); err != nil {
		t.Fatalf("Save() todo2 failed: %v", err)
	}

	// Filter by pending status
	pendingStatus := domain.StatusPending
	todos, err := repo.FindAll(context.Background(), ports.Filters{
		Status: &pendingStatus,
	})

	if err != nil {
		t.Fatalf("FindAll() unexpected error: %v", err)
	}

	if len(todos) != 1 {
		t.Errorf("FindAll() with status filter returned %d todos, want 1", len(todos))
	}

	if len(todos) > 0 && todos[0].Status() != domain.StatusPending {
		t.Errorf("FindAll() returned todo with status %v, want %v", todos[0].Status(), domain.StatusPending)
	}
}

func TestPostgresTodRepository_FindAll_WithPriorityFilter_FiltersCorrectly(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Create todos with different priorities
	title1, _ := domain.NewTaskTitle("Low Priority Todo")
	todo1 := domain.NewTodo(title1, "Description", domain.PriorityLow, nil)

	title2, _ := domain.NewTaskTitle("High Priority Todo")
	todo2 := domain.NewTodo(title2, "Description", domain.PriorityHigh, nil)

	if err := repo.Save(context.Background(), todo1); err != nil {
		t.Fatalf("Save() todo1 failed: %v", err)
	}
	if err := repo.Save(context.Background(), todo2); err != nil {
		t.Fatalf("Save() todo2 failed: %v", err)
	}

	// Filter by high priority
	highPriority := domain.PriorityHigh
	todos, err := repo.FindAll(context.Background(), ports.Filters{
		Priority: &highPriority,
	})

	if err != nil {
		t.Fatalf("FindAll() unexpected error: %v", err)
	}

	if len(todos) != 1 {
		t.Errorf("FindAll() with priority filter returned %d todos, want 1", len(todos))
	}

	if len(todos) > 0 && todos[0].Priority() != domain.PriorityHigh {
		t.Errorf("FindAll() returned todo with priority %v, want %v", todos[0].Priority(), domain.PriorityHigh)
	}
}

func TestPostgresTodRepository_FindAll_WithLimit_LimitsResults(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Create multiple todos
	for i := 0; i < 5; i++ {
		todo := createTestTodo()
		if err := repo.Save(context.Background(), todo); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
	}

	// Query with limit
	limit := 2
	todos, err := repo.FindAll(context.Background(), ports.Filters{
		Limit: &limit,
	})

	if err != nil {
		t.Fatalf("FindAll() unexpected error: %v", err)
	}

	if len(todos) != 2 {
		t.Errorf("FindAll() with limit returned %d todos, want 2", len(todos))
	}
}

func TestPostgresTodRepository_FindAll_WithOffset_OffsetsResults(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Create multiple todos
	var createdIDs []domain.TodoID
	for i := 0; i < 3; i++ {
		todo := createTestTodo()
		createdIDs = append(createdIDs, todo.ID())
		if err := repo.Save(context.Background(), todo); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
		// Small delay to ensure different created_at times
		time.Sleep(10 * time.Millisecond)
	}

	// Query with offset
	offset := 1
	todos, err := repo.FindAll(context.Background(), ports.Filters{
		Offset: &offset,
	})

	if err != nil {
		t.Fatalf("FindAll() unexpected error: %v", err)
	}

	if len(todos) != 2 {
		t.Errorf("FindAll() with offset returned %d todos, want 2", len(todos))
	}
}

func TestPostgresTodRepository_Update_ExistingTodo_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Save initial todo
	todo := createTestTodo()
	if err := repo.Save(context.Background(), todo); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Update the todo
	newTitle, _ := domain.NewTaskTitle("Updated Title")
	todo.UpdateTitle(newTitle)

	err := repo.Update(context.Background(), todo)

	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}

	// Verify update
	updated, err := repo.FindByID(context.Background(), todo.ID())
	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}

	if updated.Title().String() != "Updated Title" {
		t.Errorf("Title = %v, want %v", updated.Title().String(), "Updated Title")
	}
}

func TestPostgresTodRepository_Update_NonExistentTodo_ReturnsError(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Try to update non-existent todo
	todo := createTestTodo()
	err := repo.Update(context.Background(), todo)

	if err == nil {
		t.Error("Update() expected error for non-existent todo, got nil")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Update() error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestPostgresTodRepository_Update_CompleteTodo_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Save initial todo
	todo := createTestTodo()
	if err := repo.Save(context.Background(), todo); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Complete the todo
	todo.Complete()

	err := repo.Update(context.Background(), todo)

	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}

	// Verify status changed
	updated, err := repo.FindByID(context.Background(), todo.ID())
	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}

	if updated.Status() != domain.StatusCompleted {
		t.Errorf("Status = %v, want %v", updated.Status(), domain.StatusCompleted)
	}
}

func TestPostgresTodRepository_Delete_ExistingTodo_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Save todo
	todo := createTestTodo()
	if err := repo.Save(context.Background(), todo); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Delete todo
	err := repo.Delete(context.Background(), todo.ID())

	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	// Verify todo is deleted
	_, err = repo.FindByID(context.Background(), todo.ID())
	if err != domain.ErrTodoNotFound {
		t.Errorf("FindByID() after delete error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestPostgresTodRepository_Delete_NonExistentTodo_ReturnsError(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Try to delete non-existent todo
	nonExistentID := domain.NewTodoID()
	err := repo.Delete(context.Background(), nonExistentID)

	if err == nil {
		t.Error("Delete() expected error for non-existent todo, got nil")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Delete() error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestPostgresTodRepository_Reconstitution_PreservesAllFields(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Create todo with all fields set
	title, _ := domain.NewTaskTitle("Complete Todo")
	futureDate := time.Now().Add(48 * time.Hour)
	dueDate, _ := domain.NewDueDate(futureDate)
	todo := domain.NewTodo(title, "Full description", domain.PriorityUrgent, &dueDate)
	todo.Complete()

	// Save
	if err := repo.Save(context.Background(), todo); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Retrieve
	retrieved, err := repo.FindByID(context.Background(), todo.ID())
	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}

	// Verify all fields
	if retrieved.ID() != todo.ID() {
		t.Errorf("ID mismatch")
	}
	if retrieved.Title().String() != todo.Title().String() {
		t.Errorf("Title = %v, want %v", retrieved.Title().String(), todo.Title().String())
	}
	if retrieved.Description() != todo.Description() {
		t.Errorf("Description = %v, want %v", retrieved.Description(), todo.Description())
	}
	if retrieved.Status() != todo.Status() {
		t.Errorf("Status = %v, want %v", retrieved.Status(), todo.Status())
	}
	if retrieved.Priority() != todo.Priority() {
		t.Errorf("Priority = %v, want %v", retrieved.Priority(), todo.Priority())
	}
}

func TestPostgresTodRepository_ConcurrentSaves_Success(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostgresTodRepository(pool)

	// Create multiple todos concurrently
	const numTodos = 10
	errChan := make(chan error, numTodos)

	for i := 0; i < numTodos; i++ {
		go func() {
			todo := createTestTodo()
			errChan <- repo.Save(context.Background(), todo)
		}()
	}

	// Check all saves succeeded
	for i := 0; i < numTodos; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent save %d failed: %v", i, err)
		}
	}

	// Verify count
	todos, err := repo.FindAll(context.Background(), ports.Filters{})
	if err != nil {
		t.Fatalf("FindAll() unexpected error: %v", err)
	}

	if len(todos) != numTodos {
		t.Errorf("Expected %d todos, got %d", numTodos, len(todos))
	}
}
