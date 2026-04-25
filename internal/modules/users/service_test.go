package users

import (
	"errors"
	"testing"
	"time"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/modules/adminusers"
)

type fakeRepository struct {
	users             map[int64]*database.User
	usernameConflicts map[string]int64
	emailConflicts    map[string]int64
}

type fakeAdminRepository struct {
	adminUsers map[int64]*database.AdminUser
}

func (f *fakeRepository) FindByID(id int64) (*database.User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (f *fakeRepository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.User, int64, error) {
	items := make([]database.User, 0)
	for _, item := range f.users {
		items = append(items, *item)
	}
	return items, int64(len(items)), nil
}

func (f *fakeRepository) UsernameExists(username string, excludeID int64) (bool, error) {
	id, ok := f.usernameConflicts[username]
	return ok && id != excludeID, nil
}

func (f *fakeRepository) EmailExists(email string, excludeID int64) (bool, error) {
	id, ok := f.emailConflicts[email]
	return ok && id != excludeID, nil
}

func (f *fakeRepository) Create(user *database.User) error {
	user.ID = int64(len(f.users) + 1)
	f.users[user.ID] = user
	return nil
}

func (f *fakeRepository) Save(user *database.User) error {
	f.users[user.ID] = user
	return nil
}

func (f *fakeAdminRepository) FindByID(id int64) (*database.AdminUser, error) {
	adminUser, ok := f.adminUsers[id]
	if !ok {
		return nil, adminusers.ErrAdminUserNotFound
	}
	return adminUser, nil
}

func (f *fakeAdminRepository) List(filter adminusers.ListFilter, pagination common.PaginationQuery) ([]database.AdminUser, int64, error) {
	return nil, 0, nil
}

func (f *fakeAdminRepository) UsernameExists(username string, excludeID int64) (bool, error) {
	return false, nil
}

func (f *fakeAdminRepository) EmailExists(email string, excludeID int64) (bool, error) {
	return false, nil
}

func (f *fakeAdminRepository) Create(adminUser *database.AdminUser) error {
	return nil
}

func (f *fakeAdminRepository) Save(adminUser *database.AdminUser) error {
	return nil
}

func TestCreateRejectsDuplicateUsername(t *testing.T) {
	repo := &fakeRepository{
		users:             map[int64]*database.User{},
		usernameConflicts: map[string]int64{"existing": 2},
		emailConflicts:    map[string]int64{},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo, security.NewPasswordHasher(0))

	_, err := service.Create(1, CreateUserInput{
		Username: "existing",
		Password: "secret",
		FullName: "New User",
	})
	if !errors.Is(err, ErrUsernameAlreadyExists) {
		t.Fatalf("Create() error = %v, want %v", err, ErrUsernameAlreadyExists)
	}
}

func TestUpdateRejectsDuplicateEmail(t *testing.T) {
	repo := &fakeRepository{
		users: map[int64]*database.User{
			1: {ID: 1, Username: "user1", FullName: "User 1", IsActive: true},
		},
		usernameConflicts: map[string]int64{},
		emailConflicts:    map[string]int64{"dup@example.com": 2},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo, security.NewPasswordHasher(0))
	email := "dup@example.com"

	_, err := service.Update(1, 1, UpdateUserInput{
		Username: "user1",
		FullName: "User 1",
		Email:    &email,
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("Update() error = %v, want %v", err, ErrEmailAlreadyExists)
	}
}

func TestResetPasswordUpdatesPasswordHash(t *testing.T) {
	now := time.Now()
	repo := &fakeRepository{
		users: map[int64]*database.User{
			2: {ID: 2, Username: "user", FullName: "User", IsActive: true, PasswordHash: "old", CreatedAt: now, UpdatedAt: now},
		},
		usernameConflicts: map[string]int64{},
		emailConflicts:    map[string]int64{},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	hasher := security.NewPasswordHasher(0)
	service := NewService(repo, adminRepo, hasher)

	item, err := service.ResetPassword(1, 2, "new-secret")
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	if item.Username != "user" {
		t.Fatalf("ResetPassword() username = %q, want %q", item.Username, "user")
	}

	if repo.users[2].PasswordHash == "old" {
		t.Fatal("ResetPassword() did not update password hash")
	}

	if err := hasher.Compare(repo.users[2].PasswordHash, "new-secret"); err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
}
