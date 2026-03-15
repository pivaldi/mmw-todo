// cmd/migrate.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	dbpgcli "github.com/ovya/ogl/db/cli"
	oglmigrator "github.com/ovya/ogl/db/migrator"
	oglos "github.com/ovya/ogl/os"
	oglslog "github.com/ovya/ogl/slog"
	"github.com/pivaldi/mmw/todo/internal/infra/config"
	"github.com/pivaldi/mmw/todo/internal/infra/persistence/migrations"
	"github.com/pressly/goose/v3"

	// Needed to embed all the go migrations
	_ "github.com/pivaldi/mmw/todo/internal/infra/persistence/migrations/scripts"
	"github.com/rotisserie/eris"
)

var exit = 0

func main() {
	var db *sql.DB

	defer func() {
		if db != nil {
			db.Close()
		}

		os.Exit(exit)
	}()

	goose.SetLogger(&oglmigrator.FancyLogger{})
	goose.SetDebug(true)
	goose.SetSequential(true)

	ctx := context.Background()
	conf, err := config.Load(ctx, oglos.EnvMap())
	if err != nil {
		logError("loading config failed", err)

		return
	}

	db, err = sql.Open("postgres", conf.GetDatabaseURL())
	if err != nil {
		logError("can not open database connection", err)

		return
	}

	if err := db.PingContext(context.Background()); err != nil {
		logError("can ping database connection", err)

		return
	}

	options := []goose.OptionsFunc{
		goose.WithAllowMissing(),
	}

	m := oglmigrator.New(db, migrations.FS, "scripts", options...)

	migrateCmd := dbpgcli.NewMigrateCmd(m)

	if err := migrateCmd.Execute(); err != nil {
		logError("command failed", err)

		return
	}
}

func logError(msg string, err error) {
	exit = 1
	logger := slog.New(oglslog.StderrTxtHandler(slog.LevelDebug, nil))
	logger.Error(msg)
	// Print the formatted stack trace directly to stderr
	fmt.Fprintf(os.Stderr, "%s\n", eris.ToString(err, true))
}
