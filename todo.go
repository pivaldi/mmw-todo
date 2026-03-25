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
	"github.com/pivaldi/mmw-todo/internal/domain"
	"github.com/rotisserie/eris"
	"golang.org/x/sync/errgroup"
)

const (
	relayTableName = "todo.event"
	ModuleName     = "Auth"
)

var NotifyEvents = domain.AllEvents

type module struct {
	relay  *ogloutbox.EventsRelay
	server *oglserver.HTTPServer
	logger *slog.Logger
}

// Ensure Module implements oglcore.Module
var _ oglcore.Module = (*module)(nil)

type Infrastructure struct {
	DBPool   *pgxpool.Pool
	EventBus oglevents.SystemEventBus
	AuthSvc  defauth.AuthService
	Logger   *slog.Logger
}

func New(infra Infrastructure) (*module, error) {
	cfg, err := config.Load(context.Background(), "")
	if err != nil {
		return nil, eris.Wrap(err, "app failed to load config")
	}
	mux := http.NewServeMux()

	todoRepo := postgres.NewPostgresTodoRepository(infra.DBPool)
	eventDispatcher := events.NewPostgresOutboxDispatcher(infra.DBPool)
	todoService := application.NewTodoApplicationService(todoRepo, ogluow.New(infra.DBPool), eventDispatcher)

	path, handler := todov1connect.NewTodoServiceHandler(
		connecthandler.NewTodoHandler(todoService),
		connect.WithInterceptors(oglconnect.NewErrorLoggingInterceptor(infra.Logger)),
	)

	// Wrap Connect handler with auth middleware — every todo RPC requires a valid JWT
	mux.Handle(path, connecthandler.NewAuthMiddleware(infra.AuthSvc, infra.Logger, nil, handler))

	httpInfra := oglserver.HTTPServerInfra{
		Config:      cfg.Server,
		Handler:     mux,
		Logger:      infra.Logger,
		HealthFns:   oglserver.HealthFns{"database": todoService.Health},
		LogPayloads: true,
	}

	withDebug := cfg.Environment.IsDev()
	httpServer := oglserver.NewHTTPServer2(withDebug, httpInfra)
	// Initialize everything internal to Todo here!
	return &module{
		relay:  ogloutbox.NewEnventsRelay(infra.DBPool, infra.EventBus, infra.Logger, relayTableName),
		server: httpServer,
		logger: infra.Logger,
	}, nil
}

// Close properly releases allocated resources
// Example: If you had a local cache or an internal batch processor:
//
//	if err := m.internalCache.Flush(); err != nil {
//	    return err
//	}
func (m *module) Close() error {
	m.logger.Info("shutting down module internal resources")

	return nil
}

func (m *module) Start(ctx context.Context) error {
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
