package middleware

import (
	"sayakaya/pkg/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func TraceMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = uuid.New().String()
			}

			ctx := logger.WithTraceID(c.Request().Context(), reqID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func RequestIDConfig() middleware.RequestIDConfig {
	return middleware.RequestIDConfig{
		TargetHeader: echo.HeaderXRequestID,
	}
}
