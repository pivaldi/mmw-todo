package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	todov1connect "github.com/pivaldi/mmw/contracts/go/todo/v1/todov1connect"
	"github.com/pivaldi/mmw/todo/internal/adapters/events"
	connecthandler "github.com/pivaldi/mmw/todo/internal/adapters/handler/connect"
	"github.com/pivaldi/mmw/todo/internal/adapters/repository/postgres"
	"github.com/pivaldi/mmw/todo/internal/application"
)

// Config holds application configuration
type Config struct {
	DatabaseURL string
	Port        string
	Environment string
}

func main() {
	// Load configuration
	config := loadConfig()

	// Setup logger
	logger := setupLogger(config.Environment)

	// Start application
	if err := run(config, logger); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}

// run starts the application with proper lifecycle management
func run(config Config, logger *slog.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	logger.Info("connecting to database", "url", maskDatabaseURL(config.DatabaseURL))
	dbPool, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("creating database pool: %w", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err := dbPool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	logger.Info("database connection established")

	// Initialize dependencies (Dependency Injection)
	todoRepository := postgres.NewPostgresTodRepository(dbPool)
	eventDispatcher := events.NewInMemoryEventDispatcher(logger)
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
		Addr: ":" + config.Port,
		Handler: h2c.NewHandler(
			corsMiddleware(loggingMiddleware(mux, logger)),
			&http2.Server{},
		),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting server", "port", config.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Setup signal handling for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)

		// Create context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
		defer shutdownCancel()

		// Gracefully shut down the server
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			// Force close after timeout
			if err := server.Close(); err != nil {
				logger.Error("forcing server close", "error", err)
			}
			return fmt.Errorf("graceful shutdown: %w", err)
		}

		logger.Info("server stopped gracefully")
	}

	return nil
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5435/todoapp?sslmode=disable"),
		Port:        getEnv("PORT", "8090"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
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

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// maskDatabaseURL masks sensitive parts of database URL for logging
func maskDatabaseURL(url string) string {
	// Simple masking - in production use more robust URL parsing
	if len(url) < 20 {
		return "***"
	}
	return url[:10] + "***" + url[len(url)-10:]
}
