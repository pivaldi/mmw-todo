package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	ogldb "github.com/ovya/ogl/db"
	ogluow "github.com/ovya/ogl/pg/uow"
	"github.com/rotisserie/eris"

	"github.com/pivaldi/mmw-todo/internal/application/ports"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

// PostgresTodoRepository implements the TodoRepository port using PostgreSQL
type PostgresTodoRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTodoRepository creates a new PostgreSQL repository
func NewPostgresTodoRepository(pool *pgxpool.Pool) *PostgresTodoRepository {
	return &PostgresTodoRepository{pool: pool}
}

// Save persists a new todo to the database.
func (r *PostgresTodoRepository) Save(ctx context.Context, todo *domain.Todo) error {
	query := `
		INSERT INTO todo.todo (id, title, description, status, priority, due_date,
		                       created_at, updated_at, completed_at, user_id)
		VALUES (@id, @title, @description, @status, @priority, @due_date,
		        @created_at, @updated_at, @completed_at, @user_id)
	`
	exec := ogluow.GetExecutor(ctx, r.pool)
	_, err := exec.Exec(ctx, query, pgx.NamedArgs(ogldb.StructArgs(todo.Snapshot())))
	if err != nil {
		return eris.Wrap(err, "saving todo")
	}

	return nil
}

// BatchSave persists multiple todos efficiently in a single network trip.
// Uses positional args — pgx.Batch is incompatible with pgx.NamedArgs.
func (r *PostgresTodoRepository) BatchSave(ctx context.Context, todos []*domain.Todo) error {
	if len(todos) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO todo.todo (id, title, description, status, priority, due_date,
		                       created_at, updated_at, completed_at, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, todo := range todos {
		snap := todo.Snapshot()
		batch.Queue(query,
			snap.ID.String(),
			snap.Title,
			snap.Description,
			snap.Status,
			snap.Priority,
			snap.DueDate,
			snap.CreatedAt,
			snap.UpdatedAt,
			snap.CompletedAt,
			snap.UserID.String(),
		)
	}

	exec := ogluow.GetExecutor(ctx, r.pool)
	br := exec.SendBatch(ctx, batch)
	defer br.Close()

	for i := range todos {
		_, err := br.Exec()
		if err != nil {
			return eris.Wrapf(err, "batch insert failed at index %d", i)
		}
	}

	return nil
}

// FindByID retrieves a todo by its ID scoped to the given user.
func (r *PostgresTodoRepository) FindByID(ctx context.Context, id domain.TodoID, userID uuid.UUID) (*domain.Todo, error) {
	query := `
		SELECT id, title, description, status, priority, due_date,
		       created_at, updated_at, completed_at, user_id
		FROM todo.todo
		WHERE id = $1 AND user_id = $2
	`
	exec := ogluow.GetExecutor(ctx, r.pool)
	rows, err := exec.Query(ctx, query, id.String(), userID)
	if err != nil {
		return nil, eris.Wrap(err, "querying todo")
	}

	snap, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.TodoSnapshot])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTodoNotFound
		}

		return nil, eris.Wrap(err, "collecting todo")
	}

	return domain.ReconstituteTodo(&snap), nil
}

// FindAll retrieves todos matching the given filters.
// Uses dynamic positional args for filters — pgx.NamedArgs is incompatible
// with runtime-composed query strings.
func (r *PostgresTodoRepository) FindAll(ctx context.Context, filters ports.Filters) ([]*domain.Todo, error) {
	query := `
		SELECT id, title, description, status, priority, due_date,
		       created_at, updated_at, completed_at, user_id
		FROM todo.todo
		WHERE TRUE
	`
	args := []any{}
	argIndex := 1

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, filters.UserID.String())
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status.String())
		argIndex++
	}

	if filters.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, filters.Priority.String())
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *filters.Limit)
		argIndex++
	}

	if filters.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, *filters.Offset)
	}

	exec := ogluow.GetExecutor(ctx, r.pool)
	rows, err := exec.Query(ctx, query, args...)
	if err != nil {
		return nil, eris.Wrap(err, "querying todos")
	}

	snaps, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.TodoSnapshot])
	if err != nil {
		return nil, eris.Wrap(err, "collecting todos")
	}

	todos := make([]*domain.Todo, len(snaps))
	for i := range snaps {
		todos[i] = domain.ReconstituteTodo(&snaps[i])
	}

	return todos, nil
}

// Update updates an existing todo scoped to the given user.
func (r *PostgresTodoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	query := `
		UPDATE todo.todo
		SET title        = @title,
		    description  = @description,
		    status       = @status,
		    priority     = @priority,
		    due_date     = @due_date,
		    updated_at   = @updated_at,
		    completed_at = @completed_at
		WHERE id = @id AND user_id = @user_id
	`
	exec := ogluow.GetExecutor(ctx, r.pool)
	result, err := exec.Exec(ctx, query, pgx.NamedArgs(ogldb.StructArgs(todo.Snapshot())))
	if err != nil {
		return eris.Wrap(err, "updating todo")
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTodoNotFound
	}

	return nil
}

// Delete removes a todo from the database scoped to the given user.
func (r *PostgresTodoRepository) Delete(ctx context.Context, id domain.TodoID, userID uuid.UUID) error {
	exec := ogluow.GetExecutor(ctx, r.pool)
	result, err := exec.Exec(ctx,
		`DELETE FROM todo.todo WHERE id = $1 AND user_id = $2`,
		id.String(), userID.String(),
	)
	if err != nil {
		return eris.Wrap(err, "deleting todo")
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTodoNotFound
	}

	return nil
}

func (r *PostgresTodoRepository) Health(ctx context.Context) (any, error) {
	row := r.pool.QueryRow(ctx, "SELECT count(*) FROM todo.todo")
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, eris.Wrap(err, "scan row")
	}

	return count, nil
}
