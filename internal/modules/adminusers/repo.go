package adminusers

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
)

type Repository interface {
	FindByID(id int64) (*database.AdminUser, error)
	List(filter ListFilter, pagination common.PaginationQuery) ([]database.AdminUser, int64, error)
	UsernameExists(username string, excludeID int64) (bool, error)
	EmailExists(email string, excludeID int64) (bool, error)
	FindRoleGroupsByIDs(ids []int64) ([]database.RoleGroup, error)
	CreateWithRoleGroups(adminUser *database.AdminUser, roleGroups []database.RoleGroup) error
	SaveWithRoleGroups(adminUser *database.AdminUser, roleGroups []database.RoleGroup) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db.DB()}
}

func (r *repository) FindByID(id int64) (*database.AdminUser, error) {
	var adminUser database.AdminUser
	if err := r.db.Preload("RoleGroups").First(&adminUser, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminUserNotFound
		}

		return nil, fmt.Errorf("find admin user by id: %w", err)
	}

	return &adminUser, nil
}

func (r *repository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.AdminUser, int64, error) {
	query := r.db.Model(&database.AdminUser{})

	if filter.Keyword != "" {
		keyword := "%" + strings.ToLower(filter.Keyword) + "%"
		query = query.Where(
			"LOWER(username) LIKE ? OR LOWER(full_name) LIKE ? OR LOWER(COALESCE(email, '')) LIKE ?",
			keyword,
			keyword,
			keyword,
		)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.IsSuperAdmin != nil {
		query = query.Where("is_super_admin = ?", *filter.IsSuperAdmin)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count admin users: %w", err)
	}

	orderBy := buildOrderBy(pagination)

	var items []database.AdminUser
	if err := query.Preload("RoleGroups").Order(orderBy).Offset(pagination.Offset()).Limit(pagination.PageSize).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list admin users: %w", err)
	}

	return items, total, nil
}

func (r *repository) UsernameExists(username string, excludeID int64) (bool, error) {
	query := r.db.Model(&database.AdminUser{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("count admin users by username: %w", err)
	}

	return count > 0, nil
}

func (r *repository) EmailExists(email string, excludeID int64) (bool, error) {
	query := r.db.Model(&database.AdminUser{}).Where("email = ?", email)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("count admin users by email: %w", err)
	}

	return count > 0, nil
}

func (r *repository) FindRoleGroupsByIDs(ids []int64) ([]database.RoleGroup, error) {
	if len(ids) == 0 {
		return []database.RoleGroup{}, nil
	}

	var roleGroups []database.RoleGroup
	if err := r.db.Where("id IN ?", ids).Find(&roleGroups).Error; err != nil {
		return nil, fmt.Errorf("find role groups by ids: %w", err)
	}

	return roleGroups, nil
}

func (r *repository) CreateWithRoleGroups(adminUser *database.AdminUser, roleGroups []database.RoleGroup) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(adminUser).Error; err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}

		if err := tx.Model(adminUser).Association("RoleGroups").Replace(roleGroups); err != nil {
			return fmt.Errorf("replace admin user role groups: %w", err)
		}

		adminUser.RoleGroups = roleGroups
		return nil
	})
}

func (r *repository) SaveWithRoleGroups(adminUser *database.AdminUser, roleGroups []database.RoleGroup) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(adminUser).Error; err != nil {
			return fmt.Errorf("save admin user: %w", err)
		}

		if err := tx.Model(adminUser).Association("RoleGroups").Replace(roleGroups); err != nil {
			return fmt.Errorf("replace admin user role groups: %w", err)
		}

		adminUser.RoleGroups = roleGroups
		return nil
	})
}

func buildOrderBy(pagination common.PaginationQuery) string {
	column := "created_at"
	switch pagination.SortBy {
	case "id", "username", "full_name", "email", "created_at", "updated_at", "is_super_admin", "is_active":
		column = pagination.SortBy
	}

	return column + " " + pagination.SortOrder
}
