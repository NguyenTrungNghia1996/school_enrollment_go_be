package adminauth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/config"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/middleware"
)

const adminAuthPrefix = "/api/v1/admin/auth"

type Handler struct {
	service Service
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(app *fiber.App, cfg *config.Config, db *database.Database) {
	repo := NewRepository(db)
	passwordHasher := security.NewPasswordHasher(0)
	jwtService := security.NewAdminJWTService(cfg.Auth.AdminJWT)
	service := NewService(repo, passwordHasher, jwtService)
	handler := NewHandler(service)

	group := app.Group(adminAuthPrefix)
	group.Post("/login", handler.Login)
	group.Get("/me", middleware.RequireAdminAuth(cfg), handler.Me)
	group.Get("/permissions", middleware.RequireAdminAuth(cfg), handler.Permissions)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	result, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			return common.Error(c, fiber.StatusUnauthorized, common.NewError("INVALID_CREDENTIALS", "username or password is incorrect", nil))
		case errors.Is(err, ErrAdminInactive):
			return common.Error(c, fiber.StatusForbidden, common.NewError("ADMIN_INACTIVE", "admin account is inactive", nil))
		default:
			return common.Error(c, fiber.StatusInternalServerError, common.NewError("ADMIN_LOGIN_FAILED", "admin login failed", nil))
		}
	}

	return common.Success(c, fiber.StatusOK, "Admin login successful", result)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	claims, ok := middleware.AdminClaimsFromContext(c)
	if !ok || claims == nil {
		return common.Error(c, fiber.StatusUnauthorized, common.NewError("INVALID_TOKEN", "admin token is invalid", nil))
	}

	admin, err := h.service.Me(claims.ID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAdminInactive):
			return common.Error(c, fiber.StatusForbidden, common.NewError("ADMIN_INACTIVE", "admin account is inactive", nil))
		case errors.Is(err, ErrAdminNotFound):
			return common.Error(c, fiber.StatusUnauthorized, common.NewError("ADMIN_NOT_FOUND", "admin account was not found", nil))
		default:
			return common.Error(c, fiber.StatusInternalServerError, common.NewError("GET_ADMIN_PROFILE_FAILED", "failed to fetch admin profile", nil))
		}
	}

	return common.Success(c, fiber.StatusOK, "Fetched admin profile successfully", admin)
}

func (h *Handler) Permissions(c *fiber.Ctx) error {
	claims, ok := middleware.AdminClaimsFromContext(c)
	if !ok || claims == nil {
		return common.Error(c, fiber.StatusUnauthorized, common.NewError("INVALID_TOKEN", "admin token is invalid", nil))
	}

	permissions, err := h.service.Permissions(claims.ID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAdminInactive):
			return common.Error(c, fiber.StatusForbidden, common.NewError("ADMIN_INACTIVE", "admin account is inactive", nil))
		case errors.Is(err, ErrAdminNotFound):
			return common.Error(c, fiber.StatusUnauthorized, common.NewError("ADMIN_NOT_FOUND", "admin account was not found", nil))
		default:
			return common.Error(c, fiber.StatusInternalServerError, common.NewError("GET_ADMIN_PERMISSIONS_FAILED", "failed to fetch admin permissions", nil))
		}
	}

	return common.Success(c, fiber.StatusOK, "Fetched admin permissions successfully", permissions)
}
