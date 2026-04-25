package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fibercors "github.com/gofiber/fiber/v2/middleware/cors"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/config"
)

const (
	defaultCORSAllowHeaders = "Origin,Content-Type,Accept,Authorization"
	defaultCORSAllowMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	adminClaimsContextKey   = "admin_claims"
	userClaimsContextKey    = "user_claims"
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

func RequireAdminAuth(cfg *config.Config) fiber.Handler {
	jwtService := security.NewAdminJWTService(cfg.Auth.AdminJWT)
	return requireAuth(jwtService.ParseToken, adminClaimsContextKey, "admin")
}

func RequireUserAuth(cfg *config.Config) fiber.Handler {
	jwtService := security.NewUserJWTService(cfg.Auth.UserJWT)
	return requireAuth(jwtService.ParseToken, userClaimsContextKey, "user")
}

func AdminClaimsFromContext(c *fiber.Ctx) (*security.TokenClaims, bool) {
	claims, ok := c.Locals(adminClaimsContextKey).(*security.TokenClaims)
	return claims, ok
}

func UserClaimsFromContext(c *fiber.Ctx) (*security.TokenClaims, bool) {
	claims, ok := c.Locals(userClaimsContextKey).(*security.TokenClaims)
	return claims, ok
}

func requireAuth(parseToken func(string) (*security.TokenClaims, error), contextKey, subject string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractBearerToken(c.Get(fiber.HeaderAuthorization))
		if token == "" {
			return common.Error(c, fiber.StatusUnauthorized, common.NewError("UNAUTHORIZED", "missing bearer token", nil))
		}

		claims, err := parseToken(token)
		if err != nil {
			return common.Error(c, fiber.StatusUnauthorized, common.NewError("INVALID_TOKEN", subject+" token is invalid", nil))
		}

		c.Locals(contextKey, claims)
		return c.Next()
	}
}

func extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
}
