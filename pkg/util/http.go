package util

import "github.com/labstack/echo/v4"

func GetRequestId(c echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}
