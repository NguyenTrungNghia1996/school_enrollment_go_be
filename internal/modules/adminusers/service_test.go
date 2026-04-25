package adminusers

import (
	"errors"
	"testing"
	"time"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/common/security"
	"school_enrollment_be/internal/database"
)

type fakeRepository struct {
	adminUsers        map[int64]*database.AdminUser
	usernameConflicts map[string]int64
	emailConflicts    map[string]int64
}

func (f *fakeRepository) FindByID(id int64) (*database.AdminUser, error) {
	adminUser, ok := f.adminUsers[id]
	if !ok {
		return nil, ErrAdminUserNotFound
	}
	return adminUser, nil
}

func (f *fakeRepository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.AdminUser, int64, error) {
	items := make([]database.AdminUser, 0)
	for _, item := range f.adminUsers {
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

func (f *fakeRepository) Create(adminUser *database.AdminUser) error {
	adminUser.ID = int64(len(f.adminUsers) + 1)
	f.adminUsers[adminUser.ID] = adminUser
	return nil
}

func (f *fakeRepository) Save(adminUser *database.AdminUser) error {
	f.adminUsers[adminUser.ID] = adminUser
	return nil
}

func TestCreateRejectsDuplicateUsername(t *testing.T) {
	repo := &fakeRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "root", FullName: "Root", IsSuperAdmin: true, IsActive: true},
		},
		usernameConflicts: map[string]int64{"existing": 2},
		emailConflicts:    map[string]int64{},
	}

	service := NewService(repo, security.NewPasswordHasher(0))

	_, err := service.Create(1, CreateAdminUserInput{
		Username: "existing",
		Password: "secret",
		FullName: "New Admin",
	})
	if !errors.Is(err, ErrUsernameAlreadyExists) {
		t.Fatalf("Create() error = %v, want %v", err, ErrUsernameAlreadyExists)
	}
}

func TestUpdateRejectsSelfDeactivate(t *testing.T) {
	repo := &fakeRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "root", FullName: "Root", IsSuperAdmin: true, IsActive: true},
		},
		usernameConflicts: map[string]int64{},
		emailConflicts:    map[string]int64{},
	}

	service := NewService(repo, security.NewPasswordHasher(0))
	isActive := false

	_, err := service.Update(1, 1, UpdateAdminUserInput{
		Username: "root",
		FullName: "Root",
		IsActive: &isActive,
	})
	if !errors.Is(err, ErrCannotDeactivateSelf) {
		t.Fatalf("Update() error = %v, want %v", err, ErrCannotDeactivateSelf)
	}
}

func TestUpdateRejectsSuperAdminChangeByNonSuperAdmin(t *testing.T) {
	repo := &fakeRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "manager", FullName: "Manager", IsSuperAdmin: false, IsActive: true},
			2: {ID: 2, Username: "staff", FullName: "Staff", IsSuperAdmin: false, IsActive: true},
		},
		usernameConflicts: map[string]int64{},
		emailConflicts:    map[string]int64{},
	}

	service := NewService(repo, security.NewPasswordHasher(0))
	isSuperAdmin := true

	_, err := service.Update(1, 2, UpdateAdminUserInput{
		Username:     "staff",
		FullName:     "Staff",
		IsSuperAdmin: &isSuperAdmin,
	})
	if !errors.Is(err, ErrForbiddenSuperAdminChange) {
		t.Fatalf("Update() error = %v, want %v", err, ErrForbiddenSuperAdminChange)
	}
}

func TestUpdateRejectsDeactivateSuperAdmin(t *testing.T) {
	repo := &fakeRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "root", FullName: "Root", IsSuperAdmin: true, IsActive: true},
			2: {ID: 2, Username: "boss", FullName: "Boss", IsSuperAdmin: true, IsActive: true},
		},
		usernameConflicts: map[string]int64{},
		emailConflicts:    map[string]int64{},
	}

	service := NewService(repo, security.NewPasswordHasher(0))
	isActive := false

	_, err := service.Update(1, 2, UpdateAdminUserInput{
		Username: "boss",
		FullName: "Boss",
		IsActive: &isActive,
	})
	if !errors.Is(err, ErrCannotDeactivateSuperAdmin) {
		t.Fatalf("Update() error = %v, want %v", err, ErrCannotDeactivateSuperAdmin)
	}
}

func TestUpdateStatusRejectsDeactivateSuperAdmin(t *testing.T) {
	repo := &fakeRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "root", FullName: "Root", IsSuperAdmin: true, IsActive: true},
			2: {ID: 2, Username: "boss", FullName: "Boss", IsSuperAdmin: true, IsActive: true},
		},
		usernameConflicts: map[string]int64{},
		emailConflicts:    map[string]int64{},
	}

	service := NewService(repo, security.NewPasswordHasher(0))

	_, err := service.UpdateStatus(1, 2, false)
	if !errors.Is(err, ErrCannotDeactivateSuperAdmin) {
		t.Fatalf("UpdateStatus() error = %v, want %v", err, ErrCannotDeactivateSuperAdmin)
	}
}

func TestResetPasswordUpdatesPasswordHash(t *testing.T) {
	now := time.Now()
	repo := &fakeRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "root", FullName: "Root", IsSuperAdmin: true, IsActive: true, CreatedAt: now, UpdatedAt: now},
			2: {ID: 2, Username: "staff", FullName: "Staff", IsSuperAdmin: false, IsActive: true, PasswordHash: "old", CreatedAt: now, UpdatedAt: now},
		},
		usernameConflicts: map[string]int64{},
		emailConflicts:    map[string]int64{},
	}

	hasher := security.NewPasswordHasher(0)
	service := NewService(repo, hasher)

	item, err := service.ResetPassword(1, 2, "new-secret")
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	if item.Username != "staff" {
		t.Fatalf("ResetPassword() username = %q, want %q", item.Username, "staff")
	}

	if repo.adminUsers[2].PasswordHash == "old" {
		t.Fatal("ResetPassword() did not update password hash")
	}

	if err := hasher.Compare(repo.adminUsers[2].PasswordHash, "new-secret"); err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
}
