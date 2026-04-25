package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	fibercors "github.com/gofiber/fiber/v2/middleware/cors"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/config"
)

const (
	defaultCORSAllowHeaders = "Origin,Content-Type,Accept,Authorization"
	defaultCORSAllowMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
)

func Register(app *fiber.App, logger *slog.Logger, cfg *config.Config) {
	app.Use(fiberrecover.New(fiberrecover.Config{
		EnableStackTrace: cfg.App.Env != "production",
	}))

	app.Use(RequestLogger(logger))

	app.Use(fibercors.New(fibercors.Config{
		AllowOrigins: cfg.CORS.AllowOrigins,
		AllowHeaders: defaultCORSAllowHeaders,
		AllowMethods: defaultCORSAllowMethods,
	}))
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		logger.Info("http request",
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.IP(),
		)

		return err
	}
}

func FiberErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal server error"

		if fiberErr, ok := err.(*fiber.Error); ok {
			code = fiberErr.Code
			message = fiberErr.Message
		}

		logger.Error("request failed",
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", code,
			"error", err,
		)

		return common.Error(c, code, common.NewError("REQUEST_FAILED", message, nil))
	}
}
