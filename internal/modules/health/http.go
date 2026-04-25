package health

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"school_enrollment_be/internal/common/response"
	"school_enrollment_be/internal/config"
)

const apiPrefix = "/api/v1"

type Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func RegisterRoutes(app *fiber.App, cfg *config.Config) {
	handler := NewHandler(cfg)

	group := app.Group(apiPrefix)
	group.Get("/health", handler.Check)
}

func (h *Handler) Check(c *fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, "Service is healthy", fiber.Map{
		"status":      "ok",
		"app":         h.cfg.App.Name,
		"environment": h.cfg.App.Env,
		"time":        time.Now().In(h.cfg.App.Location),
	})
}
