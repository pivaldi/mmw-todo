package domain

import (
	"time"

	"github.com/google/uuid"
)

// TodoSnapshot is the Memento for the Todo aggregate.
// It captures the full persistent state as primitives so that
// pgx.RowToStructByName can scan DB columns directly via db tags.
// The events field is intentionally excluded — it is runtime state only.
type TodoSnapshot struct {
	ID          uuid.UUID  `db:"id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	Status      string     `db:"status"`
	Priority    string     `db:"priority"`
	DueDate     *time.Time `db:"due_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	CompletedAt *time.Time `db:"completed_at"`
	UserID      uuid.UUID  `db:"user_id"`
}
