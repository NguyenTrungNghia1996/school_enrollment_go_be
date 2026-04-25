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
	Create(adminUser *database.AdminUser) error
	Save(adminUser *database.AdminUser) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db.DB()}
}

func (r *repository) FindByID(id int64) (*database.AdminUser, error) {
	var adminUser database.AdminUser
	if err := r.db.First(&adminUser, "id = ?", id).Error; err != nil {
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
	if err := query.Order(orderBy).Offset(pagination.Offset()).Limit(pagination.PageSize).Find(&items).Error; err != nil {
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

func (r *repository) Create(adminUser *database.AdminUser) error {
	if err := r.db.Create(adminUser).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	return nil
}

func (r *repository) Save(adminUser *database.AdminUser) error {
	if err := r.db.Save(adminUser).Error; err != nil {
		return fmt.Errorf("save admin user: %w", err)
	}

	return nil
}

func buildOrderBy(pagination common.PaginationQuery) string {
	column := "created_at"
	switch pagination.SortBy {
	case "id", "username", "full_name", "email", "created_at", "updated_at", "is_super_admin", "is_active":
		column = pagination.SortBy
	}

	return column + " " + pagination.SortOrder
}
