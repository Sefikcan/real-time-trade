package server

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sefikcan/read-time-trade/pkg/config"
	"github.com/sefikcan/read-time-trade/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	echo   *echo.Echo
	cfg    *config.Config
	logger logger.Logger
}

func NewServer(cfg *config.Config, logger logger.Logger) *Server {
	return &Server{
		echo:   echo.New(),
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Run() error {
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", s.cfg.Server.Host, s.cfg.Server.Port),
		ReadTimeout:    time.Second * s.cfg.Server.ReadTimeout,
		WriteTimeout:   time.Second * s.cfg.Server.WriteTimeout,
		MaxHeaderBytes: s.cfg.Server.MaxHeaderBytes,
	}

	go func() {
		s.logger.Infof("Server is listening on PORT: %s", s.cfg.Server.Port)
		if err := s.echo.StartServer(server); err != nil {
			s.logger.Fatalf("Error starting server: ", err)
		}
	}()

	go func() {
		s.logger.Infof("Starting Debug Server on PORT: %s", s.cfg.Server.Port)
		if err := http.ListenAndServe(s.cfg.Server.Port, http.DefaultServeMux); err != nil {
			s.logger.Errorf("Error ListenAndServe: %s", err)
		}
	}()

	if err := s.MapHandlers(s.echo); err != nil {
		return err
	}

	// gracefull shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, shutdown := context.WithTimeout(context.Background(), s.cfg.Server.CtxTimeout*time.Second)
	defer shutdown()
	s.logger.Info("Server exited properly")
	return s.echo.Server.Shutdown(ctx)
}
