package adminusers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/database"
)

var (
	ErrAdminUserNotFound          = errors.New("admin user not found")
	ErrActorAdminNotFound         = errors.New("actor admin user not found")
	ErrActorAdminInactive         = errors.New("actor admin user is inactive")
	ErrUsernameAlreadyExists      = errors.New("username already exists")
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrCannotDeactivateSelf       = errors.New("cannot deactivate self")
	ErrCannotDeactivateSuperAdmin = errors.New("cannot deactivate super admin")
	ErrForbiddenSuperAdminChange  = errors.New("only super admin can change is_super_admin")
	ErrInvalidPassword            = errors.New("password is required")
	ErrInvalidUsername            = errors.New("username is required")
	ErrInvalidFullName            = errors.New("full name is required")
)

type Service interface {
	List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]AdminUserResponse, common.PaginationMeta, error)
	GetByID(actorID, id int64) (*AdminUserResponse, error)
	Create(actorID int64, input CreateAdminUserInput) (*AdminUserResponse, error)
	Update(actorID, id int64, input UpdateAdminUserInput) (*AdminUserResponse, error)
	UpdateStatus(actorID, id int64, isActive bool) (*AdminUserResponse, error)
	ResetPassword(actorID, id int64, newPassword string) (*AdminUserResponse, error)
}

type service struct {
	repo           Repository
	passwordHasher security.PasswordHasher
}

type ListFilter struct {
	Keyword      string
	IsActive     *bool
	IsSuperAdmin *bool
}

type CreateAdminUserInput struct {
	Username     string
	Password     string
	FullName     string
	Email        *string
	PhoneNumber  *string
	IsSuperAdmin *bool
	IsActive     *bool
}

type UpdateAdminUserInput struct {
	Username     string
	FullName     string
	Email        *string
	PhoneNumber  *string
	IsSuperAdmin *bool
	IsActive     *bool
}

type AdminUserResponse struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	FullName     string     `json:"full_name"`
	Email        *string    `json:"email,omitempty"`
	PhoneNumber  *string    `json:"phone_number,omitempty"`
	IsSuperAdmin bool       `json:"is_super_admin"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

func NewService(repo Repository, passwordHasher security.PasswordHasher) Service {
	return &service{
		repo:           repo,
		passwordHasher: passwordHasher,
	}
}

func (s *service) List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]AdminUserResponse, common.PaginationMeta, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, common.PaginationMeta{}, err
	}

	filter.Keyword = strings.TrimSpace(filter.Keyword)

	items, total, err := s.repo.List(filter, pagination)
	if err != nil {
		return nil, common.PaginationMeta{}, err
	}

	result := make([]AdminUserResponse, 0, len(items))
	for i := range items {
		result = append(result, toAdminUserResponse(&items[i]))
	}

	return result, common.BuildPaginationMeta(pagination, total), nil
}

func (s *service) GetByID(actorID, id int64) (*AdminUserResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	adminUser, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := toAdminUserResponse(adminUser)
	return &response, nil
}

func (s *service) Create(actorID int64, input CreateAdminUserInput) (*AdminUserResponse, error) {
	actor, err := s.requireActiveActor(actorID)
	if err != nil {
		return nil, err
	}

	username := strings.TrimSpace(input.Username)
	fullName := strings.TrimSpace(input.FullName)
	password := strings.TrimSpace(input.Password)
	email := normalizeOptionalString(input.Email)
	phoneNumber := normalizeOptionalString(input.PhoneNumber)

	if username == "" {
		return nil, ErrInvalidUsername
	}
	if fullName == "" {
		return nil, ErrInvalidFullName
	}
	if password == "" {
		return nil, ErrInvalidPassword
	}

	isSuperAdmin := false
	if input.IsSuperAdmin != nil {
		isSuperAdmin = *input.IsSuperAdmin
	}
	if isSuperAdmin && !actor.IsSuperAdmin {
		return nil, ErrForbiddenSuperAdminChange
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	if err := s.ensureUniqueFields(username, email, 0); err != nil {
		return nil, err
	}

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash admin user password: %w", err)
	}

	adminUser := &database.AdminUser{
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Email:        email,
		PhoneNumber:  phoneNumber,
		IsSuperAdmin: isSuperAdmin,
		IsActive:     isActive,
	}

	if err := s.repo.Create(adminUser); err != nil {
		return nil, err
	}

	response := toAdminUserResponse(adminUser)
	return &response, nil
}

func (s *service) Update(actorID, id int64, input UpdateAdminUserInput) (*AdminUserResponse, error) {
	actor, err := s.requireActiveActor(actorID)
	if err != nil {
		return nil, err
	}

	adminUser, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	username := strings.TrimSpace(input.Username)
	fullName := strings.TrimSpace(input.FullName)
	email := normalizeOptionalString(input.Email)
	phoneNumber := normalizeOptionalString(input.PhoneNumber)

	if username == "" {
		return nil, ErrInvalidUsername
	}
	if fullName == "" {
		return nil, ErrInvalidFullName
	}

	if input.IsActive != nil && !*input.IsActive && actor.ID == adminUser.ID {
		return nil, ErrCannotDeactivateSelf
	}

	if input.IsActive != nil && !*input.IsActive && adminUser.IsSuperAdmin {
		return nil, ErrCannotDeactivateSuperAdmin
	}

	if input.IsSuperAdmin != nil && *input.IsSuperAdmin != adminUser.IsSuperAdmin && !actor.IsSuperAdmin {
		return nil, ErrForbiddenSuperAdminChange
	}

	if err := s.ensureUniqueFields(username, email, adminUser.ID); err != nil {
		return nil, err
	}

	adminUser.Username = username
	adminUser.FullName = fullName
	adminUser.Email = email
	adminUser.PhoneNumber = phoneNumber

	if input.IsSuperAdmin != nil {
		adminUser.IsSuperAdmin = *input.IsSuperAdmin
	}

	if input.IsActive != nil {
		adminUser.IsActive = *input.IsActive
	}

	if err := s.repo.Save(adminUser); err != nil {
		return nil, err
	}

	response := toAdminUserResponse(adminUser)
	return &response, nil
}

func (s *service) UpdateStatus(actorID, id int64, isActive bool) (*AdminUserResponse, error) {
	actor, err := s.requireActiveActor(actorID)
	if err != nil {
		return nil, err
	}

	adminUser, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !isActive && actor.ID == adminUser.ID {
		return nil, ErrCannotDeactivateSelf
	}

	if !isActive && adminUser.IsSuperAdmin {
		return nil, ErrCannotDeactivateSuperAdmin
	}

	adminUser.IsActive = isActive
	if err := s.repo.Save(adminUser); err != nil {
		return nil, err
	}

	response := toAdminUserResponse(adminUser)
	return &response, nil
}

func (s *service) ResetPassword(actorID, id int64, newPassword string) (*AdminUserResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	adminUser, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		return nil, ErrInvalidPassword
	}

	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return nil, fmt.Errorf("hash admin user password: %w", err)
	}

	adminUser.PasswordHash = passwordHash
	if err := s.repo.Save(adminUser); err != nil {
		return nil, err
	}

	response := toAdminUserResponse(adminUser)
	return &response, nil
}

func (s *service) requireActiveActor(actorID int64) (*database.AdminUser, error) {
	actor, err := s.repo.FindByID(actorID)
	if err != nil {
		if errors.Is(err, ErrAdminUserNotFound) {
			return nil, ErrActorAdminNotFound
		}
		return nil, err
	}

	if !actor.IsActive {
		return nil, ErrActorAdminInactive
	}

	return actor, nil
}

func (s *service) ensureUniqueFields(username string, email *string, excludeID int64) error {
	usernameExists, err := s.repo.UsernameExists(username, excludeID)
	if err != nil {
		return err
	}
	if usernameExists {
		return ErrUsernameAlreadyExists
	}

	if email == nil {
		return nil
	}

	emailExists, err := s.repo.EmailExists(*email, excludeID)
	if err != nil {
		return err
	}
	if emailExists {
		return ErrEmailAlreadyExists
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

func toAdminUserResponse(adminUser *database.AdminUser) AdminUserResponse {
	return AdminUserResponse{
		ID:           adminUser.ID,
		Username:     adminUser.Username,
		FullName:     adminUser.FullName,
		Email:        adminUser.Email,
		PhoneNumber:  adminUser.PhoneNumber,
		IsSuperAdmin: adminUser.IsSuperAdmin,
		IsActive:     adminUser.IsActive,
		CreatedAt:    &adminUser.CreatedAt,
		UpdatedAt:    &adminUser.UpdatedAt,
	}
}
