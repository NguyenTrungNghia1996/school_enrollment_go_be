package adminauth

import (
	"errors"
	"fmt"
	"strings"

	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/database"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAdminInactive      = errors.New("admin user is inactive")
	ErrAdminNotFound      = errors.New("admin user not found")
	ErrInvalidToken       = errors.New("invalid admin token")
)

type Service interface {
	Login(username, password string) (*LoginResult, error)
	Me(token string) (*AdminProfile, error)
}

type service struct {
	repo           Repository
	passwordHasher security.PasswordHasher
	jwtService     security.AdminJWTService
}

type LoginResult struct {
	AccessToken string       `json:"access_token"`
	Admin       AdminProfile `json:"admin"`
}

type AdminProfile struct {
	ID           int64   `json:"id"`
	Username     string  `json:"username"`
	FullName     string  `json:"full_name"`
	Email        *string `json:"email,omitempty"`
	PhoneNumber  *string `json:"phone_number,omitempty"`
	IsSuperAdmin bool    `json:"is_super_admin"`
	IsActive     bool    `json:"is_active"`
}

func NewService(repo Repository, passwordHasher security.PasswordHasher, jwtService security.AdminJWTService) Service {
	return &service{
		repo:           repo,
		passwordHasher: passwordHasher,
		jwtService:     jwtService,
	}
}

func (s *service) Login(username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	admin, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	if !admin.IsActive {
		return nil, ErrAdminInactive
	}

	if err := s.passwordHasher.Compare(admin.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtService.GenerateToken(admin.ID)
	if err != nil {
		return nil, fmt.Errorf("generate admin access token: %w", err)
	}

	return &LoginResult{
		AccessToken: accessToken,
		Admin:       toAdminProfile(admin),
	}, nil
}

func (s *service) Me(token string) (*AdminProfile, error) {
	claims, err := s.jwtService.ParseToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	admin, err := s.repo.FindByID(claims.ID)
	if err != nil {
		return nil, err
	}

	if !admin.IsActive {
		return nil, ErrAdminInactive
	}

	profile := toAdminProfile(admin)
	return &profile, nil
}

func toAdminProfile(admin *database.AdminUser) AdminProfile {
	return AdminProfile{
		ID:           admin.ID,
		Username:     admin.Username,
		FullName:     admin.FullName,
		Email:        admin.Email,
		PhoneNumber:  admin.PhoneNumber,
		IsSuperAdmin: admin.IsSuperAdmin,
		IsActive:     admin.IsActive,
	}
}
