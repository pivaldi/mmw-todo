//  services/todo/module.go
package todo

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ovya/ogl/core"
	"github.com/ovya/ogl/postgres/uow"
	"github.com/pivaldi/mmw/contracts/gen/go/todo/v1/todov1connect"
	"github.com/pivaldi/mmw/todo/internal/adapters/events"
	connecthandler "github.com/pivaldi/mmw/todo/internal/adapters/handler/connect"
	"github.com/pivaldi/mmw/todo/internal/adapters/repository/postgres"
	"github.com/pivaldi/mmw/todo/internal/application"
	"github.com/pivaldi/mmw/todo/internal/infra/workers"
)

const panicMsg = "module is not bootstrapped. Use NewModule."

type Module struct {
	dbPool         *pgxpool.Pool
	relay          *workers.OutboxRelay
	handler        *connecthandler.TodoHandler
	isBootstrapped bool
	logger         *slog.Logger
}

// Ensure Module implements core.Module
var _ core.Module = (*Module)(nil)

func Build(dbPool *pgxpool.Pool, eventBus workers.SystemEventBus, logger *slog.Logger) *Module {
	// Initialize everything internal to Todo here!
	return &Module{
		dbPool:         dbPool,
		relay:          workers.NewOutboxRelay(dbPool, eventBus, logger),
		handler:        newTodoHandler(dbPool),
		logger:         logger,
		isBootstrapped: true,
	}
}

func (m *Module) Close() error {
	m.logger.Info("shutting down todo module internal resources")

	// Example: If you had a local cache or an internal batch processor:
	// if err := m.internalCache.Flush(); err != nil {
	//     return err
	// }

	return nil
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	if !m.isBootstrapped {
		panic(panicMsg)
	}

	path, handler := todov1connect.NewTodoServiceHandler(m.handler)
	mux.Handle(path, handler)
}

func (m *Module) StartWorkers(ctx context.Context) error {
	if !m.isBootstrapped {
		panic(panicMsg)
	}

	// This blocks until ctx is canceled (which happens on shutdown)
	m.relay.Start(ctx)

	return nil
}

func newTodoHandler(dbPool *pgxpool.Pool) *connecthandler.TodoHandler {
	todoRepository := postgres.NewPostgresTodoRepository(dbPool)
	eventDispatcher := events.NewPostgresOutboxDispatcher(dbPool)
	todoService := application.NewTodoApplicationService(todoRepository, uow.NewUnitOfWork(dbPool), eventDispatcher)

	return connecthandler.NewTodoHandler(todoService)
}
