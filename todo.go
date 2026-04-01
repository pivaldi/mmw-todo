// services/todo/todo.go
package todo

import (
	"context"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	pfconnect "github.com/piprim/mmw/pkg/platform/connect"
	pfcore "github.com/piprim/mmw/pkg/platform/core"
	pfoutbox "github.com/piprim/mmw/pkg/platform/db/outbox"
	pfevents "github.com/piprim/mmw/pkg/platform/events"
	pfuow "github.com/piprim/mmw/pkg/platform/pg/uow"
	pfserver "github.com/piprim/mmw/pkg/platform/server"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	"github.com/pivaldi/mmw-contracts/gen/go/todo/v1/todov1connect"
	connecthandler "github.com/pivaldi/mmw-todo/internal/adapters/inbound/connect"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/events"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/persistence/postgres"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/infra/config"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

const (
	relayTableName = "todo.event"
	ModuleName     = "Auth"
)

type Module struct {
	relay   *pfoutbox.EventsRelay
	server  *pfserver.HTTPServer
	logger  *slog.Logger
	service application.TodoService
}

// Service return the todo application service
func (m *Module) Service() application.TodoService {
	return m.service
}

// Ensure Module implements pfcore.Module
var _ pfcore.Module = (*Module)(nil)

type Infrastructure struct {
	DBPool   *pgxpool.Pool
	EventBus pfevents.SystemEventBus
	AuthSvc  defauth.AuthService
	Logger   *slog.Logger
}

func New(infra Infrastructure) (*Module, error) {
	cfg, err := config.Load(context.Background(), "")
	if err != nil {
		return nil, eris.Wrap(err, "app failed to load config")
	}
	mux := http.NewServeMux()

	uow := pfuow.New(infra.DBPool)
	todoRepo := postgres.NewPostgresTodoRepository(uow)
	eventDispatcher := events.NewPostgresOutboxDispatcher(uow)
	todoService := application.NewTodoApplicationService(todoRepo, uow, eventDispatcher)

	path, handler := todov1connect.NewTodoServiceHandler(
		connecthandler.NewTodoHandler(todoService),
		connect.WithInterceptors(pfconnect.NewErrorLoggingInterceptor(infra.Logger)),
	)

	// Wrap Connect handler with auth middleware — every todo RPC requires a valid JWT
	mux.Handle(path, connecthandler.NewAuthMiddleware(infra.AuthSvc, infra.Logger, nil, handler))

	httpInfra := pfserver.HTTPServerInfra{
		Config:          cfg.Server,
		Handler:         mux,
		Logger:          infra.Logger,
		HealthFns:       pfserver.HealthFns{"database": todoService.Health},
		LogPayloads:     true,
		WithDebugRoutes: cfg.Environment.IsDev(),
	}

	httpServer := pfserver.NewHTTPServer(httpInfra)
	// Initialize everything internal to Todo here!
	return &Module{
		relay:   pfoutbox.NewEnventsRelay(infra.DBPool, infra.EventBus, infra.Logger, relayTableName),
		server:  httpServer,
		logger:  infra.Logger,
		service: todoService,
	}, nil
}

// Close properly releases allocated resources
// Example: If you had a local cache or an internal batch processor:
//
//	if err := m.internalCache.Flush(); err != nil {
//	    return err
//	}
func (m *Module) Close() error {
	m.logger.Info("shutting down module internal resources")

	return nil
}

func (m *Module) Start(ctx context.Context) error {
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

	return eris.Wrapf(err, "%s failure", ModuleName)
}
