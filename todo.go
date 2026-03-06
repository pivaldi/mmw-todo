package todo

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pivaldi/mmw/contracts/gen/go/todo/v1/todov1connect"
	"github.com/pivaldi/mmw/todo/internal/adapters/events"
	connecthandler "github.com/pivaldi/mmw/todo/internal/adapters/handler/connect"
	"github.com/pivaldi/mmw/todo/internal/adapters/repository/postgres"
	"github.com/pivaldi/mmw/todo/internal/application"
	"github.com/pivaldi/mmw/todo/internal/infra/config"
	"github.com/rotisserie/eris"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	readHeaderTimeout    = 5 * time.Second
	idleTimeout          = 120 * time.Second
	shutdownTimeout      = 30 * time.Second
	minDatabaseURLLength = 20
)

type app struct {
	config      *config.Config
	logger      *slog.Logger
	todoService *application.TodoApplicationService
	todoHandler *connecthandler.TodoHandler
	dbPool      *pgxpool.Pool
}

func (app *app) logError(err error, msg string) error {
	out := eris.Wrap(err, msg)
	// TODO Log the error
	// See https://github.com/rotisserie/eris/blob/master/examples/logging/example.go

	return out
}

func New() *app {
	return &app{}
}

// Close handles the desroying resources.
func (a *app) Close() {
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}

func (a *app) Bootstrap(ctx context.Context, envs map[string]string) error {
	conf, err := config.Load(ctx, envs)
	if err != nil {
		return eris.Wrap(err, "app failed to load configuration")
	}
	a.config = conf
	a.logger = setupLogger(conf)

	dbUrl := a.config.Database.URL()
	a.logger.Info("connecting to database", "url", maskDatabaseURL(dbUrl))

	dbPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return a.logError(err, "creating database pool")
	}
	a.dbPool = dbPool

	if err := dbPool.Ping(ctx); err != nil {
		return a.logError(err, "pinging database")
	}
	a.logger.Info("database connection established")

	todoRepository := postgres.NewPostgresTodoRepository(dbPool)
	eventDispatcher := events.NewInMemoryEventDispatcher(a.logger)
	a.todoService = application.NewTodoApplicationService(todoRepository, eventDispatcher)
	a.todoHandler = connecthandler.NewTodoHandler(a.todoService)

	return nil
}

// Run runs the Connect server (HTTP+GRPC).
func (a *app) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	// FIX: Register handlers BEFORE starting the server
	path, handler := todov1connect.NewTodoServiceHandler(a.todoHandler)
	mux.Handle(path, handler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := a.dbPool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			a.logger.Error("database connection error")

			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","database":"up"}`)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"name": "Todo API", "status": "running"}`)
	})

	// Define middleware chain cleanly
	var rootHandler http.Handler = mux
	rootHandler = loggingMiddleware(rootHandler, a.logger)
	rootHandler = withCORS(a.config, rootHandler)

	server := &http.Server{
		Addr:    a.config.Server.Port.String(),
		Handler: h2c.NewHandler(rootHandler, &http2.Server{}),
		// Use ReadHeaderTimeout instead of ReadTimeout/WriteTimeout
		// This protects against slow headers but allows long streams
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		a.logger.Info("starting server", "port", a.config.Server.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown signal (ctx.Done) or server error
	select {
	case err := <-serverErrors:
		return a.logError(err, "server error")

	case <-ctx.Done(): // Triggered by signal.
		a.logger.Info("shutdown signal received")

		// If we use `ctx`, it is already canceled, and Shutdown will fail instantly.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("graceful shutdown failed", "error", err)
			if err := server.Close(); err != nil {
				a.logger.Error("forcing server close", "error", err)
			}

			return a.logError(err, "graceful shutdown")
		}

		a.logger.Info("server stopped gracefully")
	}

	return nil
}

// maskDatabaseURL masks sensitive parts of database URL for logging
func maskDatabaseURL(url string) string {
	// Simple masking - in production use more robust URL parsing
	if len(url) < minDatabaseURLLength {
		return "***"
	}

	return url[:10] + "***" + url[len(url)-10:]
}

// setupLogger creates a structured logger based on environment
func setupLogger(conf *config.Config) *slog.Logger {
	var handler slog.Handler

	if conf.Environment == config.EnvironmentProduction {
		// JSON format for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: conf.LogLevel.SlogLevel(),
		})
	} else {
		// Text format for development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: conf.LogLevel.SlogLevel(),
		})
	}

	return slog.New(handler)
}
