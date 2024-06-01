package pkg

import (
	"context"
	"fmt"
	"github.com/WildEgor/e-shop-cdn/internal/adapters"
	"github.com/WildEgor/e-shop-cdn/internal/adapters/ws"
	"github.com/WildEgor/e-shop-cdn/internal/configs"
	"github.com/WildEgor/e-shop-cdn/internal/db"
	mongo "github.com/WildEgor/e-shop-cdn/internal/db/mongo"
	eh "github.com/WildEgor/e-shop-cdn/internal/handlers/errors"
	nfm "github.com/WildEgor/e-shop-cdn/internal/middlewares/not_found"
	ws_middleware "github.com/WildEgor/e-shop-cdn/internal/middlewares/ws"
	"github.com/WildEgor/e-shop-cdn/internal/routers"
	"github.com/WildEgor/e-shop-cdn/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/google/wire"
	"log/slog"
	"os"
	"time"
)

// AppSet link main app deps
var AppSet = wire.NewSet(
	configs.ConfigsSet,
	db.DbSet,
	adapters.AdaptersSet,
	routers.RouterSet,
	NewApp,
)

// Server represents the main server configuration.
type Server struct {
	App       *fiber.App
	WS        *ws.Hub
	Cron      *services.CronService
	Mongo     *mongo.Connection
	AppConfig *configs.AppConfig
}

// Run start service with deps
func (srv *Server) Run(ctx context.Context) {
	slog.Info("server is listening")

	go srv.WS.Run()

	if err := srv.Cron.Run(); err != nil {
		slog.Error("unable to start cron")
		return
	}

	if err := srv.App.Listen(fmt.Sprintf(":%s", srv.AppConfig.Port)); err != nil {
		slog.Error("unable to start server")
	}
}

// Shutdown graceful shutdown
func (srv *Server) Shutdown(ctx context.Context) {
	slog.Info("shutdown service")

	srv.WS.Stop()
	srv.Mongo.Disconnect()
	srv.Cron.Stop()

	if err := srv.App.Shutdown(); err != nil {
		slog.Error("unable to shutdown server")
	}
}

func NewApp(
	ac *configs.AppConfig,
	lc *configs.LoggerConfig,

	eh *eh.ErrorsHandler,

	prr *routers.PrivateRouter,
	pbr *routers.PublicRouter,
	sr *routers.SwaggerRouter,
	wsr *routers.SocketRouter,

	mongo *mongo.Connection,
	ws *ws.Hub,
	cs *services.CronService,
) *Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lc.Level,
	}))
	if lc.IsJSON() {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: lc.Level,
		}))
	}
	slog.SetDefault(logger)

	app := fiber.New(fiber.Config{
		AppName:      ac.Name,
		ErrorHandler: eh.Handle,
		Views:        html.New("./assets", ".html"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	app.Use(cors.New(cors.Config{
		AllowHeaders: "Origin, Content-Type, Accept, Content-Length, Accept-Language, Accept-Encoding, Authorization, Connection, Access-Control-Allow-Origin",
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
	}))
	app.Use(recover.New())
	app.Use("/ws", ws_middleware.NewWS())

	prr.Setup(app)
	pbr.Setup(app)
	sr.Setup(app)
	wsr.Setup(app)

	// 404 handler
	app.Use(nfm.NewNotFound())

	return &Server{
		App:       app,
		WS:        ws,
		Cron:      cs,
		Mongo:     mongo,
		AppConfig: ac,
	}
}
