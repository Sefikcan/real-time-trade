package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	mw "github.com/sefikcan/read-time-trade/internal/middleware"
	"github.com/sefikcan/read-time-trade/pkg/metric"
	"github.com/sefikcan/read-time-trade/pkg/util"
	echoSwagger "github.com/swaggo/echo-swagger"
	"net/http"
)

func (s *Server) MapHandlers(e *echo.Echo) error {
	metrics, err := metric.CreateMetrics(s.cfg.Metric.Url, s.cfg.Metric.ServiceName)
	if err != nil {
		s.logger.Errorf("CreateMetrics error: %s", err)
	}
	s.logger.Infof("Metrics available URL: %s, ServiceName: %s", s.cfg.Metric.Url, s.cfg.Metric.ServiceName)

	middlewareManager := mw.NewMiddlewareManager(s.cfg, s.logger)
	e.Use(middlewareManager.RequestLoggerMiddleware)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderXRequestID},
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize:         1 << 10, //1kb
		DisablePrintStack: true,
		DisableStackAll:   true,
	}))
	e.Use(middleware.RequestID())
	e.Use(middlewareManager.MetricsMiddleware(metrics))
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit("2M"))
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := e.Group("/api/v1")
	health := v1.Group("/health")

	health.GET("", func(c echo.Context) error {
		s.logger.Infof("Health check RequestID: %s", util.GetRequestId(c))
		return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
	})

	return nil
}
