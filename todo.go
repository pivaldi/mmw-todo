// services/todo/todo.go
package todo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ovya/ogl/core"
	"github.com/pivaldi/mmw/todo/internal/infra/config"
	"github.com/rotisserie/eris"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
)

const (
	readHeaderTimeout    = 5 * time.Second
	idleTimeout          = 120 * time.Second
	shutdownTimeout      = 30 * time.Second
	minDatabaseURLLength = 20
)

type app struct {
	config  *config.Config
	logger  *slog.Logger
	modules []core.Module // Keep a list of all registered modules

	// todoService    *application.TodoApplicationService
	// todoHandler    *connecthandler.TodoHandler
	dbPool         *pgxpool.Pool
	isBootstrapped bool
	// pubSub         *gochannel.GoChannel
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

func (a *app) SetModules(modules []core.Module) error {
	if !a.isBootstrapped {
		return errors.New("app is not bootstraped")
	}

	a.modules = modules

	return nil
}

func (a *app) GetBdPool() (*pgxpool.Pool, error) {
	if !a.isBootstrapped {
		return nil, errors.New("app is not bootstraped")
	}

	return a.dbPool, nil
}

// Close handles the desroying resources.
func (a *app) Close() {
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}

func (a *app) GetConfig(ctx context.Context, envs map[string]string) (*config.Config, error) {
	if a.isBootstrapped {
		return a.config, nil
	}

	conf, err := config.Load(ctx, envs)
	if err != nil {
		return nil, eris.Wrap(err, "app failed to load configuration")
	}

	return conf, nil
}

func (a *app) Bootstrap(ctx context.Context, envs map[string]string) error {
	conf, err := a.GetConfig(ctx, envs)
	if err != nil {
		return err
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

	a.isBootstrapped = true

	return nil
}

func (a *app) Run(ctx context.Context) error {
	mux := http.NewServeMux()

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

	for _, m := range a.modules {
		m.RegisterRoutes(mux)
	}

	// Define middleware chain cleanly
	var rootHandler http.Handler = mux
	rootHandler = loggingMiddleware(rootHandler, a.logger)
	rootHandler = withCORS(a.config, rootHandler)

	server := &http.Server{
		Addr:              a.config.Server.Port.String(),
		Handler:           h2c.NewHandler(rootHandler, &http2.Server{}),
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}

	// 2. Create an errgroup linked to the main context
	// If ctx is canceled (Ctrl+C), the group context (gCtx) cancels.
	// If any worker returns an error, gCtx cancels, shutting everything else down safely.
	g, gCtx := errgroup.WithContext(ctx)

	// 3. Start the HTTP Server in the group
	g.Go(func() error {
		a.logger.Info("starting server", "port", a.config.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return eris.Wrap(err, "starting server failed")
		}

		return nil
	})

	// 4. Start all module background workers in the group
	for i := range a.modules {
		g.Go(func() error {
			// This will run and block. When gCtx cancels, the worker should exit gracefully.
			return a.modules[i].StartWorkers(gCtx)
		})
	}

	// 5. Wait for Shutdown Signal
	// This goroutine listens for the context cancellation and gracefully shuts down the HTTP server
	g.Go(func() error {
		<-gCtx.Done() // Triggered by Ctrl+C OR a worker failing
		a.logger.Info("initiating graceful shutdown")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	})

	// 6. Block until everything is completely shut down
	// Wait() returns the first error that caused the group to stop, if any.
	if err := g.Wait(); err != nil {
		msg := "application stopped with error"
		a.logger.Error(msg, "err", err)

		return eris.Wrap(err, msg)
	}

	a.logger.Info("application stopped gracefully")

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
