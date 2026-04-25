package menus

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

const menusPrefix = "/api/v1/admin/menus"

type Handler struct {
	service Service
}

type CreateMenuRequest struct {
	ParentID      int64   `json:"parent_id"`
	Title         string  `json:"title"`
	KeyCode       string  `json:"key_code"`
	Icon          *string `json:"icon"`
	URL           *string `json:"url"`
	PermissionBit *int32  `json:"permission_bit"`
	IsActive      *bool   `json:"is_active"`
	SortOrder     int32   `json:"sort_order"`
}

type UpdateMenuRequest struct {
	ParentID      int64   `json:"parent_id"`
	Title         string  `json:"title"`
	KeyCode       string  `json:"key_code"`
	Icon          *string `json:"icon"`
	URL           *string `json:"url"`
	PermissionBit *int32  `json:"permission_bit"`
	IsActive      *bool   `json:"is_active"`
	SortOrder     int32   `json:"sort_order"`
}

type UpdateStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func RegisterRoutes(app *fiber.App, cfg *config.Config, db *database.Database) {
	repo := NewRepository(db)
	adminRepo := adminusers.NewRepository(db)
	service := NewService(repo, adminRepo)
	handler := NewHandler(service)

	group := app.Group(menusPrefix, middleware.RequireAdminAuth(cfg))
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

	return common.Paginated(c, fiber.StatusOK, "Fetched menus successfully", items, meta)
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

	return common.Success(c, fiber.StatusOK, "Fetched menu successfully", item)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	actorID, err := currentAdminID(c)
	if err != nil {
		return err
	}

	var req CreateMenuRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.Create(actorID, CreateMenuInput{
		ParentID:      req.ParentID,
		Title:         req.Title,
		KeyCode:       req.KeyCode,
		Icon:          req.Icon,
		URL:           req.URL,
		PermissionBit: req.PermissionBit,
		IsActive:      req.IsActive,
		SortOrder:     req.SortOrder,
	})
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusCreated, "Created menu successfully", item)
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

	var req UpdateMenuRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_REQUEST", "request body is invalid", nil))
	}

	item, svcErr := h.service.Update(actorID, id, UpdateMenuInput{
		ParentID:      req.ParentID,
		Title:         req.Title,
		KeyCode:       req.KeyCode,
		Icon:          req.Icon,
		URL:           req.URL,
		PermissionBit: req.PermissionBit,
		IsActive:      req.IsActive,
		SortOrder:     req.SortOrder,
	})
	if svcErr != nil {
		return handleServiceError(c, svcErr)
	}

	return common.Success(c, fiber.StatusOK, "Updated menu successfully", item)
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

	return common.Success(c, fiber.StatusOK, "Updated menu status successfully", item)
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
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Paginate: true,
	}

	isActive, err := parseOptionalBool(c.Query("is_active"))
	if err != nil {
		return ListFilter{}, err
	}
	filter.IsActive = isActive

	parentID, err := parseOptionalInt64(c.Query("parent_id"))
	if err != nil {
		return ListFilter{}, err
	}
	filter.ParentID = parentID

	paginate, err := parseOptionalBool(c.Query("paginate"))
	if err != nil {
		return ListFilter{}, err
	}
	if paginate != nil {
		filter.Paginate = *paginate
	}

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

func parseOptionalInt64(value string) (*int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return nil, errors.New("parent_id must be an integer greater than or equal to 0")
	}

	return &parsed, nil
}

func handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrActorAdminNotFound):
		return common.Error(c, fiber.StatusUnauthorized, common.NewError("ADMIN_NOT_FOUND", "admin account was not found", nil))
	case errors.Is(err, ErrActorAdminInactive):
		return common.Error(c, fiber.StatusForbidden, common.NewError("ADMIN_INACTIVE", "admin account is inactive", nil))
	case errors.Is(err, ErrMenuNotFound):
		return common.Error(c, fiber.StatusNotFound, common.NewError("MENU_NOT_FOUND", "menu was not found", nil))
	case errors.Is(err, ErrParentMenuNotFound):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("PARENT_MENU_NOT_FOUND", "parent menu was not found", nil))
	case errors.Is(err, ErrKeyCodeExists):
		return common.Error(c, fiber.StatusConflict, common.NewError("KEY_CODE_ALREADY_EXISTS", "key_code already exists", nil))
	case errors.Is(err, ErrInvalidTitle):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_TITLE", "title is required", nil))
	case errors.Is(err, ErrInvalidKeyCode):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_KEY_CODE", "key_code is required", nil))
	case errors.Is(err, ErrInvalidParentID):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("INVALID_PARENT_ID", "parent_id must be greater than or equal to 0", nil))
	case errors.Is(err, ErrMenuCannotBeParent):
		return common.Error(c, fiber.StatusBadRequest, common.NewError("MENU_CANNOT_BE_PARENT", "menu cannot be its own parent", nil))
	default:
		return common.Error(c, fiber.StatusInternalServerError, common.NewError("MENU_OPERATION_FAILED", "menu operation failed", nil))
	}
}
