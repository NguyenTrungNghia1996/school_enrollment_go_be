package adminauth

import (
	"errors"
	"testing"
	"time"

	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/config"
	"school_enrollment_be/internal/database"
)

type fakeRepository struct {
	adminByUsername map[string]*database.AdminUser
	adminByID       map[int64]*database.AdminUser
}

func (f *fakeRepository) FindByUsername(username string) (*database.AdminUser, error) {
	admin, ok := f.adminByUsername[username]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return admin, nil
}

func (f *fakeRepository) FindByID(id int64) (*database.AdminUser, error) {
	admin, ok := f.adminByID[id]
	if !ok {
		return nil, ErrAdminNotFound
	}
	return admin, nil
}

func TestServiceLoginSuccess(t *testing.T) {
	passwordHasher := security.NewPasswordHasher(0)
	passwordHash, err := passwordHasher.Hash("admin")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	repo := &fakeRepository{
		adminByUsername: map[string]*database.AdminUser{
			"admin": {
				ID:           1,
				Username:     "admin",
				PasswordHash: passwordHash,
				FullName:     "Admin",
				IsActive:     true,
				IsSuperAdmin: true,
			},
		},
		adminByID: map[int64]*database.AdminUser{},
	}

	service := NewService(
		repo,
		passwordHasher,
		security.NewAdminJWTService(config.JWTConfig{Secret: "admin-secret", ExpiresIn: time.Hour}),
	)

	result, err := service.Login("admin", "admin")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("Login() access token is empty")
	}

	if result.Admin.Username != "admin" {
		t.Fatalf("Login() username = %q, want %q", result.Admin.Username, "admin")
	}
}

func TestServiceLoginRejectInactiveAdmin(t *testing.T) {
	passwordHasher := security.NewPasswordHasher(0)
	passwordHash, err := passwordHasher.Hash("admin")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	repo := &fakeRepository{
		adminByUsername: map[string]*database.AdminUser{
			"admin": {
				ID:           1,
				Username:     "admin",
				PasswordHash: passwordHash,
				FullName:     "Admin",
				IsActive:     false,
			},
		},
		adminByID: map[int64]*database.AdminUser{},
	}

	service := NewService(
		repo,
		passwordHasher,
		security.NewAdminJWTService(config.JWTConfig{Secret: "admin-secret", ExpiresIn: time.Hour}),
	)

	_, err = service.Login("admin", "admin")
	if !errors.Is(err, ErrAdminInactive) {
		t.Fatalf("Login() error = %v, want %v", err, ErrAdminInactive)
	}
}

func TestServiceMeRejectsInactiveAdmin(t *testing.T) {
	passwordHasher := security.NewPasswordHasher(0)
	repo := &fakeRepository{
		adminByUsername: map[string]*database.AdminUser{},
		adminByID: map[int64]*database.AdminUser{
			1: {
				ID:           1,
				Username:     "admin",
				PasswordHash: "hashed",
				FullName:     "Admin",
				IsActive:     false,
			},
		},
	}

	service := NewService(
		repo,
		passwordHasher,
		security.NewAdminJWTService(config.JWTConfig{Secret: "shared-secret", ExpiresIn: time.Hour}),
	)

	_, err := service.Me(1)
	if !errors.Is(err, ErrAdminInactive) {
		t.Fatalf("Me() error = %v, want %v", err, ErrAdminInactive)
	}
}
