package rolegroups

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/modules/adminusers"
)

var (
	ErrRoleGroupNotFound    = errors.New("role group not found")
	ErrActorAdminNotFound   = errors.New("actor admin user not found")
	ErrActorAdminInactive   = errors.New("actor admin user is inactive")
	ErrCodeAlreadyExists    = errors.New("code already exists")
	ErrInvalidName          = errors.New("name is required")
	ErrInvalidPermissionKey = errors.New("permission key is required")
)

type Service interface {
	List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]RoleGroupResponse, common.PaginationMeta, error)
	GetByID(actorID, id int64) (*RoleGroupResponse, error)
	Create(actorID int64, input CreateRoleGroupInput) (*RoleGroupResponse, error)
	Update(actorID, id int64, input UpdateRoleGroupInput) (*RoleGroupResponse, error)
	UpdateStatus(actorID, id int64, isActive bool) (*RoleGroupResponse, error)
}

type service struct {
	repo      Repository
	adminRepo adminusers.Repository
}

type ListFilter struct {
	Keyword  string
	IsActive *bool
}

type CreateRoleGroupInput struct {
	Name        string
	Description *string
	IsActive    *bool
	Permissions []PermissionInput
}

type UpdateRoleGroupInput struct {
	Name        string
	Description *string
	IsActive    *bool
	Permissions []PermissionInput
}

type PermissionInput struct {
	Key             string `json:"key"`
	PermissionValue int64  `json:"permissionValue"`
}

type RoleGroupResponse struct {
	ID          int64                `json:"id"`
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	IsActive    bool                 `json:"is_active"`
	Permissions []PermissionResponse `json:"permission"`
	CreatedAt   *time.Time           `json:"created_at,omitempty"`
	UpdatedAt   *time.Time           `json:"updated_at,omitempty"`
}

type PermissionResponse struct {
	Key             string `json:"key"`
	PermissionValue int64  `json:"permissionValue"`
}

func NewService(repo Repository, adminRepo adminusers.Repository) Service {
	return &service{repo: repo, adminRepo: adminRepo}
}

func (s *service) List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]RoleGroupResponse, common.PaginationMeta, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, common.PaginationMeta{}, err
	}

	filter.Keyword = strings.TrimSpace(filter.Keyword)
	items, total, err := s.repo.List(filter, pagination)
	if err != nil {
		return nil, common.PaginationMeta{}, err
	}

	result := make([]RoleGroupResponse, 0, len(items))
	for i := range items {
		result = append(result, toRoleGroupResponse(&items[i]))
	}

	return result, common.BuildPaginationMeta(pagination, total), nil
}

func (s *service) GetByID(actorID, id int64) (*RoleGroupResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	roleGroup, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := toRoleGroupResponse(roleGroup)
	return &response, nil
}

func (s *service) Create(actorID int64, input CreateRoleGroupInput) (*RoleGroupResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	description := normalizeOptionalString(input.Description)

	if name == "" {
		return nil, ErrInvalidName
	}

	code := generateCodeFromName(name)
	if code == "" {
		return nil, ErrInvalidName
	}

	exists, err := s.repo.CodeExists(code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCodeAlreadyExists
	}

	permissions, err := normalizePermissions(input.Permissions)
	if err != nil {
		return nil, err
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	roleGroup := &database.RoleGroup{
		Code:        code,
		Name:        name,
		Description: description,
		IsActive:    isActive,
	}

	if err := s.repo.CreateWithPermissions(roleGroup, permissions); err != nil {
		return nil, err
	}

	response := toRoleGroupResponse(roleGroup)
	return &response, nil
}

func (s *service) Update(actorID, id int64, input UpdateRoleGroupInput) (*RoleGroupResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	roleGroup, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	description := normalizeOptionalString(input.Description)

	if name == "" {
		return nil, ErrInvalidName
	}

	code := generateCodeFromName(name)
	if code == "" {
		return nil, ErrInvalidName
	}

	exists, err := s.repo.CodeExists(code, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCodeAlreadyExists
	}

	permissions, err := normalizePermissions(input.Permissions)
	if err != nil {
		return nil, err
	}

	roleGroup.Code = code
	roleGroup.Name = name
	roleGroup.Description = description
	if input.IsActive != nil {
		roleGroup.IsActive = *input.IsActive
	}

	if err := s.repo.SaveWithPermissions(roleGroup, permissions); err != nil {
		return nil, err
	}

	response := toRoleGroupResponse(roleGroup)
	return &response, nil
}

func (s *service) UpdateStatus(actorID, id int64, isActive bool) (*RoleGroupResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	roleGroup, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	roleGroup.IsActive = isActive
	if err := s.repo.SaveWithPermissions(roleGroup, roleGroup.Permissions); err != nil {
		return nil, err
	}

	response := toRoleGroupResponse(roleGroup)
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

func normalizePermissions(inputs []PermissionInput) ([]database.RoleGroupPermission, error) {
	if len(inputs) == 0 {
		return []database.RoleGroupPermission{}, nil
	}

	permissions := make([]database.RoleGroupPermission, 0, len(inputs))
	for _, input := range inputs {
		key := strings.TrimSpace(input.Key)
		if key == "" {
			return nil, ErrInvalidPermissionKey
		}

		permissions = append(permissions, database.RoleGroupPermission{
			PermissionKey:   key,
			PermissionValue: input.PermissionValue,
		})
	}

	return permissions, nil
}

func generateCodeFromName(name string) string {
	var b strings.Builder
	lastDash := false

	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case unicode.IsSpace(r) || r == '-' || r == '_':
			if b.Len() > 0 && !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	code := strings.Trim(b.String(), "-")
	if len(code) > 50 {
		code = strings.Trim(code[:50], "-")
	}

	return code
}

func toRoleGroupResponse(roleGroup *database.RoleGroup) RoleGroupResponse {
	permissions := make([]PermissionResponse, 0, len(roleGroup.Permissions))
	for _, permission := range roleGroup.Permissions {
		permissions = append(permissions, PermissionResponse{
			Key:             permission.PermissionKey,
			PermissionValue: permission.PermissionValue,
		})
	}

	return RoleGroupResponse{
		ID:          roleGroup.ID,
		Code:        roleGroup.Code,
		Name:        roleGroup.Name,
		Description: roleGroup.Description,
		IsActive:    roleGroup.IsActive,
		Permissions: permissions,
		CreatedAt:   &roleGroup.CreatedAt,
		UpdatedAt:   &roleGroup.UpdatedAt,
	}
}
