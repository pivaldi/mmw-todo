// services/todo/todo.go
package todo

import (
	"context"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	ogloutbox "github.com/ovya/ogl/db/outbox"
	ogluow "github.com/ovya/ogl/pg/uow"
	oglconnect "github.com/ovya/ogl/platform/connect"
	oglcore "github.com/ovya/ogl/platform/core"
	oglevents "github.com/ovya/ogl/platform/events"
	oglserver "github.com/ovya/ogl/platform/server"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/pivaldi/mmw-contracts/gen/go/todo/v1/todov1connect"
	"github.com/pivaldi/mmw-todo/config"
	connecthandler "github.com/pivaldi/mmw-todo/internal/adapters/inbound/connect"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/events"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/persistence/postgres"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

const relayTableName = "todo.event"
const AppName = "Auth"

type App struct {
	relay  *ogloutbox.EventsRelay
	server *oglserver.HTTPServer
	logger *slog.Logger
}

// Ensure Module implements oglcore.Module
var _ oglcore.App = (*App)(nil)

type Infrastructure struct {
	DBPool   *pgxpool.Pool
	EventBus oglevents.SystemEventBus
	AuthSvc  defauth.AuthService
	Logger   *slog.Logger
	cfg      *config.Config
}

func (i *Infrastructure) WithConfig(cfg *config.Config) Infrastructure {
	i.cfg = cfg
	return *i
}

func New(infra Infrastructure) (*App, error) {
	var cfg = infra.cfg
	if cfg == nil {
		var err error
		cfg, err = config.Load(context.Background(), "")
		if err != nil {
			return nil, eris.Wrap(err, "app failed to load config")
		}
	}
	mux := http.NewServeMux()
	path, handler := todov1connect.NewTodoServiceHandler(
		newTodoHandler(infra.DBPool),
		connect.WithInterceptors(oglconnect.NewErrorLoggingInterceptor(infra.Logger)),
	)

	// Wrap Connect handler with auth middleware — every todo RPC requires a valid JWT
	mux.Handle(path, connecthandler.NewAuthMiddleware(infra.AuthSvc, infra.Logger, nil, handler))

	// Initialize everything internal to Todo here!
	return &App{
		relay:  ogloutbox.NewEnventsRelay(infra.DBPool, infra.EventBus, infra.Logger, relayTableName),
		server: oglserver.NewHTTPServer(cfg.Environment.String(), cfg.Server, mux, infra.Logger),
		logger: infra.Logger,
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

func (m *App) Start(ctx context.Context) error {
	m.logger.Info("starting the app")
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return m.server.Start(gCtx)
	})

	if m.relay != nil {
		g.Go(func() error {
			m.relay.Start(gCtx)

			return nil
		})
	}

	err := g.Wait()

	return eris.Wrapf(err, "%s failure", AppName)
}

func newTodoHandler(dbPool *pgxpool.Pool) *connecthandler.TodoHandler {
	todoRepository := postgres.NewPostgresTodoRepository(dbPool)
	eventDispatcher := events.NewPostgresOutboxDispatcher(dbPool)
	todoService := application.NewTodoApplicationService(todoRepository, ogluow.NewUnitOfWork(dbPool), eventDispatcher)

	return connecthandler.NewTodoHandler(todoService)
}
