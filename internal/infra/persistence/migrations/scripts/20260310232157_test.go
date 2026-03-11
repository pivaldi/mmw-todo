package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits // Goose wants it
func init() {
	goose.AddMigrationContext(up20260310232157, down20260310232157)
}

func up20260310232157(ctx context.Context, tx *sql.Tx) error {
	// Implement migration logic here
	return nil
}

func down20260310232157(ctx context.Context, tx *sql.Tx) error {
	// Implement rollback logic here
	return nil
}
