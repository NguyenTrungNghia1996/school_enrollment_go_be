package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"school_enrollment_be/internal/common/response"
	"school_enrollment_be/internal/config"
	"school_enrollment_be/internal/database"
)

const apiPrefix = "/api/v1"

type Handler struct {
	cfg *config.Config
	db  *database.Database
}

func NewHandler(cfg *config.Config, db *database.Database) *Handler {
	return &Handler{cfg: cfg, db: db}
}

func RegisterRoutes(app *fiber.App, cfg *config.Config, db *database.Database) {
	handler := NewHandler(cfg, db)

	group := app.Group(apiPrefix)
	group.Get("/health", handler.Check)
}

func (h *Handler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	statusCode := fiber.StatusOK
	message := "Service is healthy"
	if err := h.db.HealthCheck(ctx); err != nil {
		dbStatus = "error"
		statusCode = fiber.StatusServiceUnavailable
		message = "Service is unhealthy"
	}

	return response.Success(c, statusCode, message, fiber.Map{
		"app": fiber.Map{
			"name":        h.cfg.App.Name,
			"environment": h.cfg.App.Env,
			"status":      "ok",
			"time":        time.Now().In(h.cfg.App.Location),
		},
		"db": fiber.Map{
			"status": dbStatus,
		},
	})
}
