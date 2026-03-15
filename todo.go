// services/todo/todo.go
package todo

import (
	"context"

	"log/slog"
	"net/http"

	"github.com/pivaldi/mmw/todo/internal/infra/config"
	"github.com/rotisserie/eris"

	"github.com/jackc/pgx/v5/pgxpool"
	ogloutbox "github.com/ovya/ogl/db/outbox"
	oglos "github.com/ovya/ogl/os"
	ogluow "github.com/ovya/ogl/pg/uow"
	oglcore "github.com/ovya/ogl/platform/core"
	oglevents "github.com/ovya/ogl/platform/events"
	oglserver "github.com/ovya/ogl/platform/server"
	"github.com/pivaldi/mmw/contracts/gen/go/todo/v1/todov1connect"
	connecthandler "github.com/pivaldi/mmw/todo/internal/adapters/inbound/connect"
	"github.com/pivaldi/mmw/todo/internal/adapters/outbound/events"
	"github.com/pivaldi/mmw/todo/internal/adapters/outbound/persistence/postgres"
	"github.com/pivaldi/mmw/todo/internal/application"
	"golang.org/x/sync/errgroup"
)

const relayTableName = "events"

type App struct {
	appName string
	relay   *ogloutbox.EventsRelay
	server  *oglserver.HTTPServer
	logger  *slog.Logger
}

// Ensure Module implements oglcore.Module
var _ oglcore.Module = (*App)(nil)

var conf *config.Config

func GetConfig(ctx context.Context, envs map[string]string) (*config.Config, error) {
	if conf != nil {
		return conf, nil
	}

	conf, err := config.Load(ctx, envs)
	if err != nil {
		return nil, eris.Wrap(err, "failed to load todo configuration")
	}

	return conf, nil
}

func New(dbPool *pgxpool.Pool, eventBus oglevents.SystemEventBus, logger *slog.Logger) (*App, error) {
	conf, err := GetConfig(context.Background(), oglos.EnvMap())
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	path, handler := todov1connect.NewTodoServiceHandler(newTodoHandler(dbPool))
	mux.Handle(path, handler)

	// Initialize everything internal to Todo here!
	return &App{
		appName: conf.GetAppName(),
		relay:   ogloutbox.NewEnventsRelay(dbPool, eventBus, logger, relayTableName),
		server:  oglserver.NewHTTPServer("todo-api", conf.GetServerPort(), mux, logger),
		logger:  logger,
	}, nil
}

// Close properly releases allocated resources
// Example: If you had a local cache or an internal batch processor:
//
//	if err := m.internalCache.Flush(); err != nil {
//	    return err
//	}
func (m *App) Close() error {
	m.logger.Info("shutting down module internal resources")

	return nil
}

func (m *App) GetName() string {
	return m.appName
}

func (m *App) Start(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return m.server.Start(gCtx)
	})

	if m.relay != nil {
		g.Go(func() error {
			// relay.Start MUST respects context cancellation
			m.relay.Start(gCtx)

			return nil
		})
	}

	return eris.Wrapf(g.Wait(), "%s failure", m.GetName())
}

func newTodoHandler(dbPool *pgxpool.Pool) *connecthandler.TodoHandler {
	todoRepository := postgres.NewPostgresTodoRepository(dbPool)
	eventDispatcher := events.NewPostgresOutboxDispatcher(dbPool)
	todoService := application.NewTodoApplicationService(todoRepository, ogluow.NewUnitOfWork(dbPool), eventDispatcher)

	return connecthandler.NewTodoHandler(todoService)
}
