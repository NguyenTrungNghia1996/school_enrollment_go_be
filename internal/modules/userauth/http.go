package userauth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/config"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/middleware"
)

const userAuthPrefix = "/api/v1/user/auth"

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
	jwtService := security.NewUserJWTService(cfg.Auth.UserJWT)
	service := NewService(repo, passwordHasher, jwtService)
	handler := NewHandler(service)

	group := app.Group(userAuthPrefix)
	group.Post("/login", handler.Login)
	group.Get("/me", middleware.RequireUserAuth(cfg), handler.Me)
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
		case errors.Is(err, ErrUserInactive):
			return common.Error(c, fiber.StatusForbidden, common.NewError("USER_INACTIVE", "user account is inactive", nil))
		default:
			return common.Error(c, fiber.StatusInternalServerError, common.NewError("USER_LOGIN_FAILED", "user login failed", nil))
		}
	}

	return common.Success(c, fiber.StatusOK, "User login successful", result)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	claims, ok := middleware.UserClaimsFromContext(c)
	if !ok || claims == nil {
		return common.Error(c, fiber.StatusUnauthorized, common.NewError("INVALID_TOKEN", "user token is invalid", nil))
	}

	user, err := h.service.Me(claims.ID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserInactive):
			return common.Error(c, fiber.StatusForbidden, common.NewError("USER_INACTIVE", "user account is inactive", nil))
		case errors.Is(err, ErrUserNotFound):
			return common.Error(c, fiber.StatusUnauthorized, common.NewError("USER_NOT_FOUND", "user account was not found", nil))
		default:
			return common.Error(c, fiber.StatusInternalServerError, common.NewError("GET_USER_PROFILE_FAILED", "failed to fetch user profile", nil))
		}
	}

	return common.Success(c, fiber.StatusOK, "Fetched user profile successfully", user)
}
