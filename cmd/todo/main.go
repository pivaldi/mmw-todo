package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ovya/ogl/platform"
	oglcore "github.com/ovya/ogl/platform/core"
	oglevents "github.com/ovya/ogl/platform/events"
	oglslog "github.com/ovya/ogl/slog"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/pivaldi/mmw-contracts/gen/go/auth/v1/authv1connect"
	todo "github.com/pivaldi/mmw-todo"
	"github.com/pivaldi/mmw-todo/internal/infra/config"
	"github.com/rotisserie/eris"
)

const (
	outputChannelBufferSize = 1024
	minDatabaseURLLength    = 20
)

var errFormater = eris.ToJSON

var (
	logger   *slog.Logger
	dbPool   *pgxpool.Pool
	exitCode = 0
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer func() {
		cancel()
		dbPool.Close()
		os.Exit(exitCode)
	}()

	var err error

	todoConf, err := config.Load(ctx, "")
	if err != nil {
		exitCode = 1
		_, _ = fmt.Fprint(os.Stdout, eris.ToString(err, true)+"\n")

		return
	}

	logger, err = oglslog.New(oglslog.HandlerText, todoConf.LogLevel.SlogLevel())
	if err != nil {
		exitCode = 1
		_, _ = fmt.Fprint(os.Stdout, eris.ToString(err, true)+"\n")

		return
	}

	todoLogger := logger.With("module", todo.ModuleName)

	watermillLogger := watermill.NewSlogLogger(todoLogger)
	rawBus := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: outputChannelBufferSize,
			// Persistent guarantees the channel won't drop messages if no subscriber is attached yet
			Persistent: true,
		},
		watermillLogger,
	)

	defer rawBus.Close()
	// Wrap the raw infrastructure in the Adapter.
	systemBus := oglevents.NewWatermillBus(rawBus)

	// When extracted, you might swap Watermill's GoChannel for RabbitMQ here!
	// systemBus := setupRabbitMQ()
	// Wrap the raw infrastructure in the Adapter.
	// systemBus := oglevents.NewWatermillBus(rawBus)

	dbPool, err = getDatabasePoolConnexion(ctx, todoLogger, todoConf.Database.URL())
	if err != nil {
		logError("creating database pool", err)

		return
	}

	// authGrpc := authv1connect.NewAuthServiceClient(httpClient connect.HTTPClient)
	authHttpClient := authv1connect.NewAuthServiceClient(
		&http.Client{}, // no TLS needed for localhost
		todoConf.AuthServer.URL("", nil),
	)

	todoModule, err := todo.New(todo.Infrastructure{
		DBPool:   dbPool,
		EventBus: systemBus,
		Logger:   todoLogger,
		AuthSvc:  defauth.NewHttpClient(authHttpClient), // defauth.NewInprocClient(authModule)
	})
	if err != nil {
		logError("creating module failed", err)
		return
	}

	// notifLogger := logger.With("module", "notifications")
	modules := []oglcore.Module{
		todoModule,
		// Use RabitMQ consummer instead
		// notifications.Build(rawBus, notifLogger),
	}

	err = platform.New(logger, modules).Run(ctx)
	if err != nil {
		logError("platform error", err)
		exitCode = 1
	}
}

func logError(msg string, err error) {
	l := slog.New(oglslog.StderrTxtHandler(slog.LevelDebug, nil))
	l.Error(msg, "details", errFormater(err, true))
}

func getDatabasePoolConnexion(ctx context.Context, logger *slog.Logger, dbUrl string) (*pgxpool.Pool, error) {
	logger.Info("connecting to database", "url", maskDatabaseURL(dbUrl))

	dbPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, eris.Wrap(err, "connecting to database")
	}

	if err := dbPool.Ping(ctx); err != nil {
		return dbPool, eris.Wrap(err, "pinging database")
	}

	logger.Info("database connection established")

	return dbPool, nil
}

// maskDatabaseURL masks sensitive parts of database URL for logging
func maskDatabaseURL(url string) string {
	// Simple masking - in production use more robust URL parsing
	if len(url) < minDatabaseURLLength {
		return "***"
	}

	return url[:10] + "***" + url[len(url)-10:]
}
