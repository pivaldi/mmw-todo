package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits // Goose wants it
func init() {
	goose.AddMigrationContext(up20260310162056, down20260310162056)
}

func up20260310162056(_ context.Context, _ *sql.Tx) error {
	return nil
}

func down20260310162056(_ context.Context, _ *sql.Tx) error {
	return nil
}
