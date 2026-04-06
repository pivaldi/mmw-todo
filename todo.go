// services/todo/todo.go
package todo

import (
	"context"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	pfconnect "github.com/piprim/mmw/pkg/platform/connect"
	pfcore "github.com/piprim/mmw/pkg/platform/core"
	pfdbmigrator "github.com/piprim/mmw/pkg/platform/db/migrator"
	pfoutbox "github.com/piprim/mmw/pkg/platform/db/outbox"
	pfevents "github.com/piprim/mmw/pkg/platform/events"
	pfuow "github.com/piprim/mmw/pkg/platform/pg/uow"
	pfserver "github.com/piprim/mmw/pkg/platform/server"
	defauth "github.com/pivaldi/mmw-contracts/definitions/auth"
	tododef "github.com/pivaldi/mmw-contracts/definitions/todo"
	"github.com/pivaldi/mmw-contracts/gen/go/todo/v1/todov1connect"
	connecthandler "github.com/pivaldi/mmw-todo/internal/adapters/inbound/connect"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/events"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/persistence/postgres"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/infra/config"
	"github.com/pivaldi/mmw-todo/internal/infra/persistence/migrations"
	"github.com/rotisserie/eris"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
)

const (
	relayTableName = "todo.event"
	ModuleName     = "Todo"
	PGSchema       = "todo"
)

type Module struct {
	relay   *pfoutbox.EventsRelay
	server  *pfserver.HTTPServer
	logger  *slog.Logger
	service application.TodoService
}

// Service returns the todo service as a tododef.TodoService, wrapped in a
// ContractAdapter so callers receive the proto-typed contract interface.
func (m *Module) Service() tododef.TodoService {
	return application.NewContractAdapter(m.service)
}

// Handler returns the module's HTTP handler so tests can wrap it in
// httptest.NewServer without starting a real server on a port.
func (m *Module) Handler() http.Handler {
	return m.server.Handler()
}

// Migrate runs all pending database migrations for the auth module.
// Intended for use in tests and migration tooling.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	m, err := pfdbmigrator.New(db, migrations.FS, "scripts", PGSchema)
	if err != nil {
		return eris.Wrap(err, "failed to create migrator")
	}

	_, err = m.Up(ctx)

	return eris.Wrap(err, "failed to migrate up")
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

	h2cHandler := h2c.NewHandler(mux, &http2.Server{})
	httpInfra := pfserver.HTTPServerInfra{
		Config:          cfg.Server,
		Handler:         h2cHandler,
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
