package rolegroups

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/config"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/middleware"
	"school_enrollment_be/internal/modules/adminusers"
)

const roleGroupsPrefix = "/api/v1/admin/role-groups"

type Handler struct {
	service Service
}

type CreateRoleGroupRequest struct {
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	IsActive    *bool               `json:"is_active"`
	Permissions []PermissionRequest `json:"permission"`
}

type UpdateRoleGroupRequest struct {
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	IsActive    *bool               `json:"is_active"`
	Permissions []PermissionRequest `json:"permission"`
}

type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type PermissionRequest struct {
	Key             string `json:"key"`
	PermissionValue int64  `json:"permissionValue"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(app *fiber.App, cfg *config.Config, db *database.Database) {
	repo := NewRepository(db)
	adminRepo := adminusers.NewRepository(db)
	service := NewService(repo, adminRepo)
	handler := NewHandler(service)

	group := app.Group(roleGroupsPrefix, middleware.RequireAdminAuth(cfg))
	group.Get("", handler.List)
	group.Get("/:id", handler.GetByID)
	group.Post("", handler.Create)
	group.Put("/:id", handler.Update)
	group.Patch("/:id/status", handler.UpdateStatus)
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

	return common.Paginated(c, fiber.StatusOK, "Fetched role groups successfully", items, meta)
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

	return common.Success(c, fiber.StatusOK, "Fetched role group successfully", item)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	var req CreateRoleGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.Create(actorID, CreateRoleGroupInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Permissions: toPermissionInputs(req.Permissions),
	})
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusCreated, "Created role group successfully", item)
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

	var req UpdateRoleGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.Update(actorID, id, UpdateRoleGroupInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Permissions: toPermissionInputs(req.Permissions),
	})
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusOK, "Updated role group successfully", item)
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

	return common.Success(c, fiber.StatusOK, "Updated role group status successfully", item)
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
	case errors.Is(err, ErrRoleGroupNotFound):
		return common.Error(c, fiber.StatusNotFound, common.NewError("ROLE_GROUP_NOT_FOUND", "role group was not found", nil))
	case errors.Is(err, ErrCodeAlreadyExists):
		return common.Error(c, fiber.StatusConflict, common.NewError("CODE_ALREADY_EXISTS", "code already exists", nil))
	case errors.Is(err, ErrInvalidName):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_NAME", "name is required", nil))
	case errors.Is(err, ErrInvalidPermissionKey):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_PERMISSION_KEY", "permission key is required", nil))
	default:
		return common.Error(c, fiber.StatusInternalServerError, common.NewError("ROLE_GROUP_OPERATION_FAILED", "role group operation failed", nil))
	}
}

func toPermissionInputs(requests []PermissionRequest) []PermissionInput {
	if len(requests) == 0 {
		return nil
	}

	items := make([]PermissionInput, 0, len(requests))
	for _, request := range requests {
		items = append(items, PermissionInput{
			Key:             request.Key,
			PermissionValue: request.PermissionValue,
		})
	}

	return items
}
