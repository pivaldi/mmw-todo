// services/todo/todo.go
package todo

import (
	"context"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	pfconnect "github.com/piprim/mmw/pkg/platform/connect"
	pfcore "github.com/piprim/mmw/pkg/platform/core"
	pfdbmigrator "github.com/piprim/mmw/pkg/platform/db/migrator"
	pfoutbox "github.com/piprim/mmw/pkg/platform/db/outbox"
	pfevents "github.com/piprim/mmw/pkg/platform/events"
	pfmiddleware "github.com/piprim/mmw/pkg/platform/middleware"
	pfuow "github.com/piprim/mmw/pkg/platform/pg/uow"
	pfserver "github.com/piprim/mmw/pkg/platform/server"
	defauth "github.com/pivaldi/mmw-contracts/go/application/auth"
	"github.com/pivaldi/mmw-contracts/go/network/todo/v1/todov1connect"
	connecthandler "github.com/pivaldi/mmw-todo/internal/adapters/inbound/connect"
	inevents "github.com/pivaldi/mmw-todo/internal/adapters/inbound/events"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/events"
	"github.com/pivaldi/mmw-todo/internal/adapters/outbound/persistence/postgres"
	"github.com/pivaldi/mmw-todo/internal/application"
	"github.com/pivaldi/mmw-todo/internal/application/command"
	"github.com/pivaldi/mmw-todo/internal/infra/config"
	"github.com/pivaldi/mmw-todo/internal/infra/persistence/migrations"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

const (
	relayTableName = "todo.event"
	ModuleName     = "Todo"
	PGSchema       = "todo"
)

type Module struct {
	relay   *pfoutbox.EventsRelay   // m.relay.Start(gCtx)  ← outbox relay (DB → bus)
	server  *pfserver.HTTPServer    // m.server.Start(gCtx) ← HTTP server
	router  *message.Router         // m.router.Run(gCtx)   ← Watermill router (bus → handlers)
	logger  *slog.Logger            // The native Go logger is enough
	service application.TodoService // This is an interface ! Can be replaced by any implementation.
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
	DBPool     *pgxpool.Pool
	EventBus   pfevents.SystemEventBus
	Subscriber message.Subscriber
	AuthSvc    defauth.AuthPrivateService
	Logger     *slog.Logger
}

// New wires all the dependencies of the Todo module and returns a ready-to-start Module.
func New(infra Infrastructure) (*Module, error) {
	cfg, err := config.Load(context.Background(), "")
	if err != nil {
		return nil, eris.Wrap(err, "app failed to load config")
	}

	todoService := newApplicationService(infra)
	router, err := newEventRouter(infra)
	if err != nil {
		return nil, err
	}

	httpServer := newHTTPServer(cfg, infra, todoService)

	return &Module{
		// Outbox relay: polls todo.event every 2 s and forwards rows to the SystemEventBus.
		relay:   pfoutbox.NewEnventsRelay(infra.DBPool, infra.EventBus, infra.Logger, relayTableName),
		server:  httpServer,
		router:  router,
		logger:  infra.Logger,
		service: todoService,
	}, nil
}

// newApplicationService builds the infrastructure adapters (repository, outbox dispatcher,
// unit of work) and wires them into the TodoApplicationService.
//
// The UnitOfWork is the single source of truth for database access: both the repository
// and the event dispatcher receive the same UoW so that writes to todo rows and writes
// to the outbox table share the same transaction.
func newApplicationService(infra Infrastructure) application.TodoService {
	uow := pfuow.New(infra.DBPool)
	todoRepo := postgres.NewPostgresTodoRepository(uow)
	eventDispatcher := events.NewPostgresOutboxDispatcher(uow)

	return application.NewTodoApplicationService(todoRepo, uow, eventDispatcher)
}

// newEventRouter creates the Watermill message router and registers all inbound event
// handlers for the Todo module.
//
// Currently the only subscription is "auth.user.deleted.v1": when a user account is
// deleted the auth module publishes that event, and this handler removes all of the
// user's tasks to keep the database clean.
func newEventRouter(infra Infrastructure) (*message.Router, error) {
	watermillLogger := watermill.NewSlogLogger(infra.Logger)

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create event router")
	}

	// Re-create the event dispatcher here so the handler's command has its own UoW
	// instance — the router runs in a separate goroutine and must not share
	// the service's UoW.
	uow := pfuow.New(infra.DBPool)
	eventDispatcher := events.NewPostgresOutboxDispatcher(uow)
	deleteUserTasksCmd := command.NewDeleteUserTasksCommand(eventDispatcher)

	router.AddConsumerHandler(
		"todo.on_auth_user_deleted", // unique handler name (must be stable across restarts)
		defauth.TopicUserDeleted,    // source topic published by the auth module
		infra.Subscriber,            // underlying pub/sub transport (GoChannel, NATS, …)
		inevents.HandleUserDeleted(deleteUserTasksCmd),
	)

	return router, nil
}

// newHTTPServer mounts the Connect RPC handler on an HTTP mux and wraps it with
// platform middleware, then returns a pre-configured HTTPServer ready to be started.
//
// Auth middleware is applied to every route: all Todo RPCs require a valid JWT.
// The service's Health method is exposed at GET /debug/monit so the platform runner
// can probe database connectivity.
// gRPC server reflection is enabled so grpcui can discover the service schema without
// a compiled proto descriptor.
func newHTTPServer(cfg *config.Config, infra Infrastructure, todoService application.TodoService) *pfserver.HTTPServer {
	mux := http.NewServeMux()

	// Register the Connect RPC handler with an error-logging interceptor.
	path, handler := todov1connect.NewTodoServiceHandler(
		connecthandler.NewTodoHandler(todoService),
		connect.WithInterceptors(pfconnect.NewErrorLoggingInterceptor(infra.Logger)),
	)

	// Every Todo RPC requires a valid JWT — the TokenValidator calls the auth module's
	// private service to validate the token and extract the user UUID.
	authMiddleware := pfmiddleware.BearerAuthMiddleware(
		connecthandler.NewTokenValidator(infra.AuthSvc),
		infra.Logger,
		nil, // no excluded paths — all routes are protected
	)
	mux.Handle(path, authMiddleware(handler))

	return pfserver.NewHTTPServer(pfserver.HTTPServerInfra{
		Config:       cfg.Server,
		Handler:      mux,
		Logger:       infra.Logger,
		HealthFns:    pfserver.HealthFns{"database": todoService.Health},
		LogPayloads:  true,
		ServiceNames: []string{todov1connect.TodoServiceName}, // enables gRPC reflection
	})
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

	// Package errgroup provides synchronization, error propagation, and Context
	// cancellation for groups of goroutines working on subtasks of a common task.
	g, gCtx := errgroup.WithContext(ctx)

	// Start the HTTP server
	g.Go(func() error {
		return m.server.Start(gCtx)
	})

	// Start the Outbox relay
	if m.relay != nil {
		g.Go(func() error {
			m.relay.Start(gCtx)

			return nil
		})
	}

	// Start the Watermill message router triggering Todo module handlers for inbound event handlers.
	g.Go(func() error {
		return m.router.Run(gCtx)
	})

	// Wait until the context is cancled or a goroutine returns an error or panics.
	err := g.Wait()

	return eris.Wrapf(err, "%s failure", ModuleName)
}
