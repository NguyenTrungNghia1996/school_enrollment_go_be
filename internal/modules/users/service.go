package users

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/modules/adminusers"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrActorAdminNotFound    = errors.New("actor admin user not found")
	ErrActorAdminInactive    = errors.New("actor admin user is inactive")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrInvalidPassword       = errors.New("password is required")
	ErrInvalidUsername       = errors.New("username is required")
	ErrInvalidFullName       = errors.New("full name is required")
)

type Service interface {
	List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]UserResponse, common.PaginationMeta, error)
	GetByID(actorID, id int64) (*UserResponse, error)
	Create(actorID int64, input CreateUserInput) (*UserResponse, error)
	Update(actorID, id int64, input UpdateUserInput) (*UserResponse, error)
	UpdateStatus(actorID, id int64, isActive bool) (*UserResponse, error)
	ResetPassword(actorID, id int64, newPassword string) (*UserResponse, error)
}

type service struct {
	repo           Repository
	adminRepo      adminusers.Repository
	passwordHasher security.PasswordHasher
}

type ListFilter struct {
	Keyword  string
	IsActive *bool
}

type CreateUserInput struct {
	Username    string
	Password    string
	FullName    string
	Email       *string
	PhoneNumber *string
	IsActive    *bool
}

type UpdateUserInput struct {
	Username    string
	FullName    string
	Email       *string
	PhoneNumber *string
	IsActive    *bool
}

type UserResponse struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	FullName    string     `json:"full_name"`
	Email       *string    `json:"email,omitempty"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func NewService(repo Repository, adminRepo adminusers.Repository, passwordHasher security.PasswordHasher) Service {
	return &service{
		repo:           repo,
		adminRepo:      adminRepo,
		passwordHasher: passwordHasher,
	}
}

func (s *service) List(actorID int64, filter ListFilter, pagination common.PaginationQuery) ([]UserResponse, common.PaginationMeta, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, common.PaginationMeta{}, err
	}

	filter.Keyword = strings.TrimSpace(filter.Keyword)

	items, total, err := s.repo.List(filter, pagination)
	if err != nil {
		return nil, common.PaginationMeta{}, err
	}

	result := make([]UserResponse, 0, len(items))
	for i := range items {
		result = append(result, toUserResponse(&items[i]))
	}

	return result, common.BuildPaginationMeta(pagination, total), nil
}

func (s *service) GetByID(actorID, id int64) (*UserResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *service) Create(actorID int64, input CreateUserInput) (*UserResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
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

	if err := s.ensureUniqueFields(username, email, 0); err != nil {
		return nil, err
	}

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash user password: %w", err)
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	user := &database.User{
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Email:        email,
		PhoneNumber:  phoneNumber,
		IsActive:     isActive,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *service) Update(actorID, id int64, input UpdateUserInput) (*UserResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(id)
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

	if err := s.ensureUniqueFields(username, email, user.ID); err != nil {
		return nil, err
	}

	user.Username = username
	user.FullName = fullName
	user.Email = email
	user.PhoneNumber = phoneNumber
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *service) UpdateStatus(actorID, id int64, isActive bool) (*UserResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	user.IsActive = isActive
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	response := toUserResponse(user)
	return &response, nil
}

func (s *service) ResetPassword(actorID, id int64, newPassword string) (*UserResponse, error) {
	if _, err := s.requireActiveActor(actorID); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		return nil, ErrInvalidPassword
	}

	passwordHash, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return nil, fmt.Errorf("hash user password: %w", err)
	}

	user.PasswordHash = passwordHash
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	response := toUserResponse(user)
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

func toUserResponse(user *database.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsActive:    user.IsActive,
		CreatedAt:   &user.CreatedAt,
		UpdatedAt:   &user.UpdatedAt,
	}
}
