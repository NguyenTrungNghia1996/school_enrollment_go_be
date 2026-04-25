package menus

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/modules/adminusers"
)

var (
	ErrMenuNotFound       = errors.New("menu not found")
	ErrActorAdminNotFound = errors.New("actor admin user not found")
	ErrActorAdminInactive = errors.New("actor admin user is inactive")
	ErrKeyCodeExists      = errors.New("key_code already exists")
	ErrInvalidTitle       = errors.New("title is required")
	ErrInvalidKeyCode     = errors.New("key_code is required")
	ErrInvalidParentID    = errors.New("parent_id must be greater than or equal to 0")
	ErrParentMenuNotFound = errors.New("parent menu not found")
	ErrMenuCannotBeParent = errors.New("menu cannot be its own parent")
)

type Service interface {
	List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]MenuResponse, common.PaginationMeta, error)
	GetByID(actorID, id int64) (*MenuResponse, error)
	Create(actorID int64, input CreateMenuInput) (*MenuResponse, error)
	Update(actorID, id int64, input UpdateMenuInput) (*MenuResponse, error)
	UpdateStatus(actorID, id int64, isActive bool) (*MenuResponse, error)
}

type service struct {
	repo      Repository
	adminRepo adminusers.Repository
}

type ListFilter struct {
	Keyword  string
	IsActive *bool
	ParentID *int64
	Paginate bool
}

type CreateMenuInput struct {
	ParentID      int64
	Title         string
	KeyCode       string
	Icon          *string
	URL           *string
	PermissionBit *int32
	IsActive      *bool
	SortOrder     int32
}

type UpdateMenuInput struct {
	ParentID      int64
	Title         string
	KeyCode       string
	Icon          *string
	URL           *string
	PermissionBit *int32
	IsActive      *bool
	SortOrder     int32
}

type MenuResponse struct {
	ID            int64      `json:"id"`
	ParentID      int64      `json:"parent_id"`
	Title         string     `json:"title"`
	KeyCode       string     `json:"key_code"`
	Icon          *string    `json:"icon,omitempty"`
	URL           *string    `json:"url,omitempty"`
	PermissionBit *int32     `json:"permission_bit,omitempty"`
	IsActive      bool       `json:"is_active"`
	SortOrder     int32      `json:"sort_order"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

func NewService(repo Repository, adminRepo adminusers.Repository) Service {
	return &service{repo: repo, adminRepo: adminRepo}
}

func (s *service) List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]MenuResponse, common.PaginationMeta, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, common.PaginationMeta{}, err
	}

	filter.Keyword = strings.TrimSpace(filter.Keyword)
	items, total, err := s.repo.List(filter, pagination)
	if err != nil {
		return nil, common.PaginationMeta{}, err
	}

	result := make([]MenuResponse, 0, len(items))
	for i := range items {
		result = append(result, toMenuResponse(&items[i]))
	}

	meta := common.BuildPaginationMeta(pagination, total)
	if !filter.Paginate {
		meta.Page = 1
		meta.PageSize = len(result)
		meta.TotalItems = total
		if total > 0 {
			meta.TotalPages = 1
		}
	}

	return result, meta, nil
}

func (s *service) GetByID(actorID, id int64) (*MenuResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	menu, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := toMenuResponse(menu)
	return &response, nil
}

func (s *service) Create(actorID int64, input CreateMenuInput) (*MenuResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	title := strings.TrimSpace(input.Title)
	keyCode := strings.TrimSpace(input.KeyCode)
	icon := normalizeOptionalString(input.Icon)
	url := normalizeOptionalString(input.URL)

	if title == "" {
		return nil, ErrInvalidTitle
	}
	if keyCode == "" {
		return nil, ErrInvalidKeyCode
	}
	if input.ParentID < 0 {
		return nil, ErrInvalidParentID
	}

	if err := s.ensureParentExists(input.ParentID, 0); err != nil {
		return nil, err
	}

	exists, err := s.repo.KeyCodeExists(keyCode, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrKeyCodeExists
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	menu := &database.Menu{
		ParentID:      input.ParentID,
		Title:         title,
		KeyCode:       keyCode,
		Icon:          icon,
		URL:           url,
		PermissionBit: input.PermissionBit,
		IsActive:      isActive,
		SortOrder:     input.SortOrder,
	}

	if err := s.repo.Create(menu); err != nil {
		return nil, err
	}

	response := toMenuResponse(menu)
	return &response, nil
}

func (s *service) Update(actorID, id int64, input UpdateMenuInput) (*MenuResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	menu, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(input.Title)
	keyCode := strings.TrimSpace(input.KeyCode)
	icon := normalizeOptionalString(input.Icon)
	url := normalizeOptionalString(input.URL)

	if title == "" {
		return nil, ErrInvalidTitle
	}
	if keyCode == "" {
		return nil, ErrInvalidKeyCode
	}
	if input.ParentID < 0 {
		return nil, ErrInvalidParentID
	}
	if input.ParentID == id {
		return nil, ErrMenuCannotBeParent
	}

	if err := s.ensureParentExists(input.ParentID, id); err != nil {
		return nil, err
	}

	exists, err := s.repo.KeyCodeExists(keyCode, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrKeyCodeExists
	}

	menu.ParentID = input.ParentID
	menu.Title = title
	menu.KeyCode = keyCode
	menu.Icon = icon
	menu.URL = url
	menu.PermissionBit = input.PermissionBit
	menu.SortOrder = input.SortOrder
	if input.IsActive != nil {
		menu.IsActive = *input.IsActive
	}

	if err := s.repo.Save(menu); err != nil {
		return nil, err
	}

	response := toMenuResponse(menu)
	return &response, nil
}

func (s *service) UpdateStatus(actorID, id int64, isActive bool) (*MenuResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	menu, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	menu.IsActive = isActive
	if err := s.repo.Save(menu); err != nil {
		return nil, err
	}

	response := toMenuResponse(menu)
	return &response, nil
}

func (s *service) requireActiveActor(actorID int64) (*database.AdminUser, error) {
	actor, err := s.adminRepo.FindByID(actorID)
	if err != nil {
		if errors.Is(err, adminusers.ErrAdminUserNotFound) {
			return nil, ErrActorAdminNotFound
		}
		return nil, err
	}

	if !actor.IsActive {
		return nil, ErrActorAdminInactive
	}

	return actor, nil
}

func (s *service) ensureParentExists(parentID, currentID int64) error {
	if parentID == 0 {
		return nil
	}
	if parentID == currentID {
		return ErrMenuCannotBeParent
	}

	_, err := s.repo.FindByID(parentID)
	if err != nil {
		if errors.Is(err, ErrMenuNotFound) {
			return ErrParentMenuNotFound
		}
		return fmt.Errorf("find parent menu: %w", err)
	}

	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func toMenuResponse(menu *database.Menu) MenuResponse {
	return MenuResponse{
		ID:            menu.ID,
		ParentID:      menu.ParentID,
		Title:         menu.Title,
		KeyCode:       menu.KeyCode,
		Icon:          menu.Icon,
		URL:           menu.URL,
		PermissionBit: menu.PermissionBit,
		IsActive:      menu.IsActive,
		SortOrder:     menu.SortOrder,
		CreatedAt:     &menu.CreatedAt,
		UpdatedAt:     &menu.UpdatedAt,
	}
}
