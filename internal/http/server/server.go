package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/config"
	appHttp "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/http"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type RouteRegisterer interface {
	RegisterRoute(router *echo.Echo)
}

type ServerConfig struct {
	Logger       *logging.Logger
	AppConfig    config.Config
	RouterConfig appHttp.RouterConfig
	ListenPort   string
}

type Server struct {
	logger     *logging.Logger
	listenPort string
	router     *echo.Echo
	httpSrv    *http.Server
}

func NewServer(cfg ServerConfig) (*Server, error) {
	err := appHttp.BuildRoutes(cfg.AppConfig, cfg.RouterConfig)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:     cfg.Logger,
		listenPort: cfg.ListenPort,
		router:     cfg.RouterConfig.MainRouter,
		httpSrv: &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.ListenPort),
			Handler: cfg.RouterConfig.MainRouter,
		},
	}, nil
}

func (srv *Server) RegisterRoute(registerer RouteRegisterer) {
	registerer.RegisterRoute(srv.router)
}

func (srv *Server) Start() {
	srv.logger.Info(fmt.Sprintf("Starting HTTP server on port %s", srv.listenPort))
	err := srv.httpSrv.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			srv.logger.Error(fmt.Sprintf("HTTP server shutdown error: %s", err.Error()))
		}
	}

	srv.logger.Info("HTTP server stopped")
}

func (srv *Server) Stop() {
	err := srv.httpSrv.Shutdown(context.Background())
	if err != nil {
		srv.logger.Error(fmt.Sprintf("Stopping server error: %s", err.Error()))
	}
}
