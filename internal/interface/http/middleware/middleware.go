package middleware

import (
	"github.com/gqadonis/colima-manager/internal/pkg/logger"
	"github.com/labstack/echo/v4"
)

func RequestLogger(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			log.Info("Request: %s %s", req.Method, req.URL.Path)
			return next(c)
		}
	}
}
