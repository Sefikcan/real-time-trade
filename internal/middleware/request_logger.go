package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/sefikcan/read-time-trade/pkg/util"
	"time"
)

func (mw *MiddlewareManager) RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)

		req := c.Request()
		res := c.Response()
		status := res.Status
		size := res.Size
		s := time.Since(start).String()
		requestId := util.GetRequestId(c)

		mw.logger.Infof("RequestId: %s, Method: %s, Url: %s, Status: %v, Size: %v, Time: %s", requestId, req.Method, req.URL, status, size, s)

		return err
	}
}
