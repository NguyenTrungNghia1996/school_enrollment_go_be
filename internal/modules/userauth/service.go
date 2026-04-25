package userauth

import (
	"errors"
	"fmt"
	"strings"

	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/database"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid user token")
)

type Service interface {
	Login(username, password string) (*LoginResult, error)
	Me(id int64) (*UserProfile, error)
}

type service struct {
	repo           Repository
	passwordHasher security.PasswordHasher
	jwtService     security.UserJWTService
}

type LoginResult struct {
	AccessToken string      `json:"access_token"`
	User        UserProfile `json:"user"`
}

type UserProfile struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	FullName    string  `json:"full_name"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	IsActive    bool    `json:"is_active"`
}

func NewService(repo Repository, passwordHasher security.PasswordHasher, jwtService security.UserJWTService) Service {
	return &service{
		repo:           repo,
		passwordHasher: passwordHasher,
		jwtService:     jwtService,
	}
}

func (s *service) Login(username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := s.passwordHasher.Compare(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate user access token: %w", err)
	}

	return &LoginResult{
		AccessToken: accessToken,
		User:        toUserProfile(user),
	}, nil
}

func (s *service) Me(id int64) (*UserProfile, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	profile := toUserProfile(user)
	return &profile, nil
}

func toUserProfile(user *database.User) UserProfile {
	return UserProfile{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsActive:    user.IsActive,
	}
}
