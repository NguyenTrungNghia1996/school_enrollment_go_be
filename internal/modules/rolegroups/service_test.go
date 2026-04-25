package rolegroups

import (
	"errors"
	"testing"
	"time"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/modules/adminusers"
)

type fakeRepository struct {
	roleGroups    map[int64]*database.RoleGroup
	codeConflicts map[string]int64
}

type fakeAdminRepository struct {
	adminUsers map[int64]*database.AdminUser
}

func (f *fakeRepository) FindByID(id int64) (*database.RoleGroup, error) {
	roleGroup, ok := f.roleGroups[id]
	if !ok {
		return nil, ErrRoleGroupNotFound
	}
	return roleGroup, nil
}

func (f *fakeRepository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.RoleGroup, int64, error) {
	items := make([]database.RoleGroup, 0)
	for _, item := range f.roleGroups {
		items = append(items, *item)
	}
	return items, int64(len(items)), nil
}

func (f *fakeRepository) CodeExists(code string, excludeID int64) (bool, error) {
	id, ok := f.codeConflicts[code]
	return ok && id != excludeID, nil
}

func (f *fakeRepository) CreateWithPermissions(roleGroup *database.RoleGroup, permissions []database.RoleGroupPermission) error {
	roleGroup.ID = int64(len(f.roleGroups) + 1)
	roleGroup.Permissions = permissions
	f.roleGroups[roleGroup.ID] = roleGroup
	return nil
}

func (f *fakeRepository) SaveWithPermissions(roleGroup *database.RoleGroup, permissions []database.RoleGroupPermission) error {
	roleGroup.Permissions = permissions
	f.roleGroups[roleGroup.ID] = roleGroup
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

func TestCreateRejectsDuplicateCode(t *testing.T) {
	repo := &fakeRepository{
		roleGroups:    map[int64]*database.RoleGroup{},
		codeConflicts: map[string]int64{"administrator": 2},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo)
	_, err := service.Create(1, CreateRoleGroupInput{
		Name: "Administrator",
	})
	if !errors.Is(err, ErrCodeAlreadyExists) {
		t.Fatalf("Create() error = %v, want %v", err, ErrCodeAlreadyExists)
	}
}

func TestUpdateRejectsEmptyName(t *testing.T) {
	now := time.Now()
	repo := &fakeRepository{
		roleGroups: map[int64]*database.RoleGroup{
			1: {ID: 1, Code: "admin", Name: "Administrator", IsActive: true, CreatedAt: now, UpdatedAt: now},
		},
		codeConflicts: map[string]int64{},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo)
	_, err := service.Update(1, 1, UpdateRoleGroupInput{
		Name: "",
	})
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("Update() error = %v, want %v", err, ErrInvalidName)
	}
}

func TestUpdateStatusSuccess(t *testing.T) {
	now := time.Now()
	repo := &fakeRepository{
		roleGroups: map[int64]*database.RoleGroup{
			1: {ID: 1, Code: "admin", Name: "Administrator", IsActive: true, CreatedAt: now, UpdatedAt: now},
		},
		codeConflicts: map[string]int64{},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo)
	item, err := service.UpdateStatus(1, 1, false)
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if item.IsActive {
		t.Fatal("UpdateStatus() is_active = true, want false")
	}
}

func TestCreatePersistsPermissions(t *testing.T) {
	repo := &fakeRepository{
		roleGroups:    map[int64]*database.RoleGroup{},
		codeConflicts: map[string]int64{},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo)
	item, err := service.Create(1, CreateRoleGroupInput{
		Name: "Administrator",
		Permissions: []PermissionInput{
			{Key: "menus", PermissionValue: 7},
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(item.Permissions) != 1 {
		t.Fatalf("permissions len = %d, want %d", len(item.Permissions), 1)
	}

	if item.Permissions[0].Key != "menus" || item.Permissions[0].PermissionValue != 7 {
		t.Fatalf("permission = %+v, want key=%q value=%d", item.Permissions[0], "menus", 7)
	}
}
