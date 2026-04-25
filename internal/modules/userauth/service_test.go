package userauth

import (
	"errors"
	"testing"
	"time"

	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/config"
	"school_enrollment_be/internal/database"
)

type fakeRepository struct {
	usersByUsername map[string]*database.User
	usersByID       map[int64]*database.User
}

func (f *fakeRepository) FindByUsername(username string) (*database.User, error) {
	user, ok := f.usersByUsername[username]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (f *fakeRepository) FindByID(id int64) (*database.User, error) {
	user, ok := f.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func TestServiceLoginSuccess(t *testing.T) {
	passwordHasher := security.NewPasswordHasher(0)
	passwordHash, err := passwordHasher.Hash("user")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	repo := &fakeRepository{
		usersByUsername: map[string]*database.User{
			"user": {
				ID:           1,
				Username:     "user",
				PasswordHash: passwordHash,
				FullName:     "Default User",
				IsActive:     true,
			},
		},
		usersByID: map[int64]*database.User{},
	}

	service := NewService(
		repo,
		passwordHasher,
		security.NewUserJWTService(config.JWTConfig{Secret: "user-secret", ExpiresIn: time.Hour}),
	)

	result, err := service.Login("user", "user")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("Login() access token is empty")
	}

	if result.User.Username != "user" {
		t.Fatalf("Login() username = %q, want %q", result.User.Username, "user")
	}
}

func TestServiceLoginRejectInactiveUser(t *testing.T) {
	passwordHasher := security.NewPasswordHasher(0)
	passwordHash, err := passwordHasher.Hash("user")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	repo := &fakeRepository{
		usersByUsername: map[string]*database.User{
			"user": {
				ID:           1,
				Username:     "user",
				PasswordHash: passwordHash,
				FullName:     "Default User",
				IsActive:     false,
			},
		},
		usersByID: map[int64]*database.User{},
	}

	service := NewService(
		repo,
		passwordHasher,
		security.NewUserJWTService(config.JWTConfig{Secret: "user-secret", ExpiresIn: time.Hour}),
	)

	_, err = service.Login("user", "user")
	if !errors.Is(err, ErrUserInactive) {
		t.Fatalf("Login() error = %v, want %v", err, ErrUserInactive)
	}
}

func TestServiceMeRejectsInactiveUser(t *testing.T) {
	passwordHasher := security.NewPasswordHasher(0)
	repo := &fakeRepository{
		usersByUsername: map[string]*database.User{},
		usersByID: map[int64]*database.User{
			1: {
				ID:           1,
				Username:     "user",
				PasswordHash: "hashed",
				FullName:     "Default User",
				IsActive:     false,
			},
		},
	}

	service := NewService(
		repo,
		passwordHasher,
		security.NewUserJWTService(config.JWTConfig{Secret: "shared-secret", ExpiresIn: time.Hour}),
	)

	_, err := service.Me(1)
	if !errors.Is(err, ErrUserInactive) {
		t.Fatalf("Me() error = %v, want %v", err, ErrUserInactive)
	}
}
