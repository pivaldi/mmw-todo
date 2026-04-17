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
	"github.com/piprim/mmw/pkg/platform"
	pfcore "github.com/piprim/mmw/pkg/platform/core"
	pfevents "github.com/piprim/mmw/pkg/platform/events"
	pfpg "github.com/piprim/mmw/pkg/platform/pg"
	pfslog "github.com/piprim/mmw/pkg/platform/slog"
	defauth "github.com/pivaldi/mmw-contracts/go/application/auth"
	"github.com/pivaldi/mmw-contracts/go/network/auth/v1/authv1connect"
	todo "github.com/pivaldi/mmw-todo"
	"github.com/pivaldi/mmw-todo/internal/infra/config"
	"github.com/rotisserie/eris"
)

const (
	outputChannelBufferSize = 1024
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

	logger, err = pfslog.New(pfslog.HandlerText, todoConf.LogLevel.SlogLevel())
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
	systemBus := pfevents.NewWatermillBus(rawBus)

	// When extracted, you might swap Watermill's GoChannel for RabbitMQ here!
	// systemBus := setupRabbitMQ()

	dbPool, err = pfpg.GetPgxPool(ctx, logger, todoConf.Database.URL())
	if err != nil {
		logError("creating database pool", err)

		return
	}

	// authGrpc := authv1connect.NewAuthPrivateServiceClient(httpClient connect.HTTPClient)
	authHTTPClient := authv1connect.NewAuthPrivateServiceClient(
		&http.Client{}, // no TLS needed for localhost
		todoConf.AuthServer.URL("", nil),
	)

	todoModule, err := todo.New(todo.Infrastructure{
		DBPool:     dbPool,
		EventBus:   systemBus,
		Subscriber: rawBus,
		Logger:     todoLogger,
		AuthSvc:    defauth.NewPrivateHTTPClient(authHTTPClient),
	})
	if err != nil {
		logError("creating module failed", err)

		return
	}

	// notifLogger := logger.With("module", "notifications")
	modules := []pfcore.Module{
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
	l := slog.New(pfslog.StderrTxtHandler(slog.LevelDebug, nil))
	l.Error(msg, "details", errFormater(err, true))
}
