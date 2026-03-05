package todo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	serverReadTimeout    = 10 * time.Second
	serverWriteTimeout   = 10 * time.Second
	serverIdleTimeout    = 120 * time.Second
	shutdownTimeout      = 30 * time.Second
	minDatabaseURLLength = 20
)

type app struct {
	config *config.Config
	logger *slog.Logger
}

func (app *app) logError(err error, msg string) error {
	out := eris.Wrap(err, msg)
	// TODO Log the error
	// See https://github.com/rotisserie/eris/blob/master/examples/logging/example.go

	return out
}

func New(envs map[string]string) (*app, error) {
	ctx := context.Background()
	conf, err := config.Load(ctx, envs)
	if err != nil {
		return nil, eris.Wrap(err, "app failed to load configuration")
	}

	app := &app{
		config: conf,
		logger: setupLogger(conf.Environment),
	}

	return app, nil
}

func (app *app) Run() error {
	if app == nil || app.config == nil {
		return errors.New("the app is not bootstraped")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	dbUrl := app.config.Database.URL()
	app.logger.Info("connecting to database", "url", maskDatabaseURL(dbUrl))
	dbPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return app.logError(err, "creating database pool")
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(ctx); err != nil {
		return app.logError(err, "pinging database")
	}
	app.logger.Info("database connection established")

	// Initialize dependencies (Dependency Injection)
	todoRepository := postgres.NewPostgresTodoRepository(dbPool)
	eventDispatcher := events.NewInMemoryEventDispatcher(app.logger)
	todoService := application.NewTodoApplicationService(todoRepository, eventDispatcher)
	todoHandler := connecthandler.NewTodoHandler(todoService)

	// Setup HTTP server with Connect handlers
	mux := http.NewServeMux()

	// Register Connect handler
	path, handler := todov1connect.NewTodoServiceHandler(todoHandler)
	mux.Handle(path, handler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		if err := dbPool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"unhealthy","database":"down"}`)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","database":"up"}`)
	})

	// Root endpoint with API information
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "name": "Todo API",
  "version": "1.0.0",
  "endpoints": {
    "health": "/health",
    "api": "/todo.v1.TodoService/*"
  },
  "protocols": ["Connect", "gRPC", "gRPC-Web"]
}`)
	})

	// Create HTTP server with h2c support (HTTP/2 without TLS for development)
	// In production, use proper TLS
	server := &http.Server{
		Addr: app.config.Server.Port.String(),
		Handler: h2c.NewHandler(
			corsMiddleware(loggingMiddleware(mux, app.logger)),
			&http2.Server{},
		),
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		IdleTimeout:  serverIdleTimeout,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		app.logger.Info("starting server", "port", app.config.Server.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Setup signal handling for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		return app.logError(err, "server error")

	case sig := <-shutdown:
		app.logger.Info("shutdown signal received", "signal", sig)

		// Create context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
		defer shutdownCancel()

		// Gracefully shut down the server
		if err := server.Shutdown(shutdownCtx); err != nil {
			app.logger.Error("graceful shutdown failed", "error", err)
			// Force close after timeout
			if err := server.Close(); err != nil {
				app.logger.Error("forcing server close", "error", err)
			}

			return app.logError(err, "graceful shutdown")
		}

		app.logger.Info("server stopped gracefully")
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
func setupLogger(environment string) *slog.Logger {
	var handler slog.Handler

	if environment == "production" {
		// JSON format for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Text format for development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// corsMiddleware adds CORS headers for development
// In production, configure more restrictive CORS policies
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version, Connect-Timeout-Ms")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
