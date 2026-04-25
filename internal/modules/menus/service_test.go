package menus

import (
	"errors"
	"testing"
	"time"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
	"school_enrollment_be/internal/modules/adminusers"
)

type fakeRepository struct {
	menus            map[int64]*database.Menu
	keyCodeConflicts map[string]int64
}

type fakeAdminRepository struct {
	adminUsers map[int64]*database.AdminUser
}

func (f *fakeRepository) FindByID(id int64) (*database.Menu, error) {
	menu, ok := f.menus[id]
	if !ok {
		return nil, ErrMenuNotFound
	}
	return menu, nil
}

func (f *fakeRepository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.Menu, int64, error) {
	items := make([]database.Menu, 0)
	for _, item := range f.menus {
		items = append(items, *item)
	}
	return items, int64(len(items)), nil
}

func (f *fakeRepository) KeyCodeExists(keyCode string, excludeID int64) (bool, error) {
	id, ok := f.keyCodeConflicts[keyCode]
	return ok && id != excludeID, nil
}

func (f *fakeRepository) Create(menu *database.Menu) error {
	menu.ID = int64(len(f.menus) + 1)
	f.menus[menu.ID] = menu
	return nil
}

func (f *fakeRepository) Save(menu *database.Menu) error {
	f.menus[menu.ID] = menu
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

func TestCreateRejectsDuplicateKeyCode(t *testing.T) {
	repo := &fakeRepository{
		menus:            map[int64]*database.Menu{},
		keyCodeConflicts: map[string]int64{"dashboard": 2},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo)
	_, err := service.Create(1, CreateMenuInput{
		Title:     "Dashboard",
		KeyCode:   "dashboard",
		ParentID:  0,
		SortOrder: 1,
	})
	if !errors.Is(err, ErrKeyCodeExists) {
		t.Fatalf("Create() error = %v, want %v", err, ErrKeyCodeExists)
	}
}

func TestUpdateRejectsSelfParent(t *testing.T) {
	now := time.Now()
	repo := &fakeRepository{
		menus: map[int64]*database.Menu{
			1: {ID: 1, ParentID: 0, Title: "Root", KeyCode: "root", IsActive: true, SortOrder: 1, CreatedAt: now, UpdatedAt: now},
		},
		keyCodeConflicts: map[string]int64{},
	}
	adminRepo := &fakeAdminRepository{
		adminUsers: map[int64]*database.AdminUser{
			1: {ID: 1, Username: "admin", FullName: "Admin", IsActive: true},
		},
	}

	service := NewService(repo, adminRepo)
	_, err := service.Update(1, 1, UpdateMenuInput{
		ParentID:  1,
		Title:     "Root",
		KeyCode:   "root",
		SortOrder: 1,
	})
	if !errors.Is(err, ErrMenuCannotBeParent) {
		t.Fatalf("Update() error = %v, want %v", err, ErrMenuCannotBeParent)
	}
}

func TestUpdateStatusSuccess(t *testing.T) {
	now := time.Now()
	repo := &fakeRepository{
		menus: map[int64]*database.Menu{
			1: {ID: 1, ParentID: 0, Title: "Root", KeyCode: "root", IsActive: true, SortOrder: 1, CreatedAt: now, UpdatedAt: now},
		},
		keyCodeConflicts: map[string]int64{},
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
