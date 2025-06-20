package app

import (
	"context"
	"os"

	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/http/server"
	masterDataHttp "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/master_data/http"
	notifHttp "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/modules/notification/http"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type App struct {
	logger *logging.Logger
	ic     *internalConnections
	srv    *server.Server
	mods   *modules
}

func New(cfg config.Config) (*App, error) {
	logger := logging.New(logging.NewHandler(os.Stderr, logging.WithAddSource(func(lvl logging.Level) bool {
		return false
	})))

	ic, err := newInternalConnections(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Run migration
	if err := RunSQLiteDatabaseMigration(ic.sqliteDb.DB); err != nil {
		return nil, err
	}

	repoColl := newRepositoryCollection(ic.db, ic.sqliteDb)
	svcColl, err := newServiceCollection(cfg, repoColl)
	if err != nil {
		return nil, err
	}

	routerConfig, err := newHttpRouterConfig(logger, svcColl)
	if err != nil {
		return nil, err
	}

	// Modules
	moduleLogger := logging.New(logging.NewHandler(os.Stderr, logging.WithAddSource(func(lvl logging.Level) bool {
		return lvl == logging.LevelError
	})))

	mods, err := newModules(cfg, ic.db, logger)
	if err != nil {
		return nil, err
	}

	srv, err := server.NewServer(server.ServerConfig{
		Logger:       logger,
		ListenPort:   cfg.HTTPPort,
		AppConfig:    cfg,
		RouterConfig: routerConfig,
	})

	if err != nil {
		return nil, err
	}

	// Notification HTTP handler
	notifHandler := notifHttp.NewHandler(cfg, moduleLogger, mods.notifProvider)
	srv.RegisterRoute(notifHandler)

	// Master data HTTP handler
	masterDataHandler := masterDataHttp.NewHandler(cfg, moduleLogger, svcColl.masterDataSvc)
	srv.RegisterRoute(masterDataHandler)

	return &App{
		logger: logger,
		ic:     ic,
		srv:    srv,
		mods:   mods,
	}, nil
}

func (app *App) Start(ctx context.Context) {
	app.srv.Start()
}

func (app *App) Stop(ctx context.Context) {
	app.ic.close()
	app.srv.Stop()
	app.mods.cleanup()
}
