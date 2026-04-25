package users

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/config"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/middleware"
	"school_enrollment_be/internal/modules/adminusers"
)

const usersPrefix = "/api/v1/admin/users"

type Handler struct {
	service Service
}

type CreateUserRequest struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	FullName    string  `json:"full_name"`
	Email       *string `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Username    string  `json:"username"`
	FullName    string  `json:"full_name"`
	Email       *string `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(app *fiber.App, cfg *config.Config, db *database.Database) {
	repo := NewRepository(db)
	adminRepo := adminusers.NewRepository(db)
	passwordHasher := security.NewPasswordHasher(0)
	service := NewService(repo, adminRepo, passwordHasher)
	handler := NewHandler(service)

	group := app.Group(usersPrefix, middleware.RequireAdminAuth(cfg))
	group.Get("", handler.List)
	group.Get("/:id", handler.GetByID)
	group.Post("", handler.Create)
	group.Put("/:id", handler.Update)
	group.Patch("/:id/status", handler.UpdateStatus)
	group.Patch("/:id/reset-password", handler.ResetPassword)
}

func (h *Handler) List(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	pagination, apiErr := common.ParsePagination(c)
	if apiErr != nil {
		return common.Error(c, fiber.StatusBadRequest, *apiErr)
	}

	filter, err := parseListFilter(c)
	if err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_FILTER", err.Error(), nil))
	}

	items, meta, svcErr := h.service.List(actorID, filter, pagination)
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Paginated(c, fiber.StatusOK, "Fetched users successfully", items, meta)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	id, parseErr := parseIDParam(c)
	if parseErr != nil {
		return parseErr
	}

	item, svcErr := h.service.GetByID(actorID, id)
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusOK, "Fetched user successfully", item)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.Create(actorID, CreateUserInput{
		Username:    req.Username,
		Password:    req.Password,
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		IsActive:    req.IsActive,
	})
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusCreated, "Created user successfully", item)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	id, parseErr := parseIDParam(c)
	if parseErr != nil {
		return parseErr
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.Update(actorID, id, UpdateUserInput{
		Username:    req.Username,
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		IsActive:    req.IsActive,
	})
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusOK, "Updated user successfully", item)
}

func (h *Handler) UpdateStatus(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	id, parseErr := parseIDParam(c)
	if parseErr != nil {
		return parseErr
	}

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.UpdateStatus(actorID, id, req.IsActive)
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusOK, "Updated user status successfully", item)
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	id, parseErr := parseIDParam(c)
	if parseErr != nil {
		return parseErr
	}

	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.ResetPassword(actorID, id, req.NewPassword)
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusOK, "Reset user password successfully", item)
}

func currentAdminID(c *fiber.Ctx) (int64, error) {
	claims, ok := middleware.AdminClaimsFromContext(c)
	if !ok || claims == nil || claims.ID <= 0 {
		return 0, common.Error(c, fiber.StatusUnauthorized, common.NewError("INVALID_TOKEN", "admin token is invalid", nil))
	}

	return claims.ID, nil
}

func parseIDParam(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id <= 0 {
		return 0, common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_ID", "id must be a positive integer", nil))
	}

	return id, nil
}

func parseListFilter(c *fiber.Ctx) (ListFilter, error) {
	filter := ListFilter{
		Keyword: strings.TrimSpace(c.Query("keyword")),
	}

	isActive, err := parseOptionalBool(c.Query("is_active"))
	if err != nil {
		return ListFilter{}, err
	}
	filter.IsActive = isActive

	return filter, nil
}

func parseOptionalBool(value string) (*bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, errors.New("boolean filter must be true or false")
	}

	return &parsed, nil
}

func handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrActorAdminNotFound):
		return common.Error(c, fiber.StatusUnauthorized, common.NewError("ADMIN_NOT_FOUND", "admin account was not found", nil))
	case errors.Is(err, ErrActorAdminInactive):
		return common.Error(c, fiber.StatusForbidden, common.NewError("ADMIN_INACTIVE", "admin account is inactive", nil))
	case errors.Is(err, ErrUserNotFound):
		return common.Error(c, fiber.StatusNotFound, common.NewError("USER_NOT_FOUND", "user was not found", nil))
	case errors.Is(err, ErrUsernameAlreadyExists):
		return common.Error(c, fiber.StatusConflict, common.NewError("USERNAME_ALREADY_EXISTS", "username already exists", nil))
	case errors.Is(err, ErrEmailAlreadyExists):
		return common.Error(c, fiber.StatusConflict, common.NewError("EMAIL_ALREADY_EXISTS", "email already exists", nil))
	case errors.Is(err, ErrInvalidPassword):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_PASSWORD", "password is required", nil))
	case errors.Is(err, ErrInvalidUsername):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_USERNAME", "username is required", nil))
	case errors.Is(err, ErrInvalidFullName):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_FULL_NAME", "full_name is required", nil))
	default:
		return common.Error(c, fiber.StatusInternalServerError, common.NewError("USER_OPERATION_FAILED", "user operation failed", nil))
	}
}
