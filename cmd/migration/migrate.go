// cmd/migrate.go
package main

import (
	"fmt"
	"log/slog"
	"os"

	pfdbcli "github.com/piprim/mmw/platform/db/cli"
	pfslog "github.com/piprim/mmw/platform/slog"

	"github.com/rotisserie/eris"
)

func main() {
	conf, err := loadConfig()
	if err != nil {
		logError("loading config failed", err)
		os.Exit(1)
	}

	if err := pfdbcli.Migrate(conf.Database.URL(), "todo", migrationsFS); err != nil {
		logError("command failed", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func logError(msg string, err error) {
	logger := slog.New(pfslog.StderrTxtHandler(slog.LevelDebug, nil))
	logger.Error(msg)
	// Print the formatted stack trace directly to stderr
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", eris.ToString(err, true))
}
