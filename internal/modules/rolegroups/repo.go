package rolegroups

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
)

type Repository interface {
	FindByID(id int64) (*database.RoleGroup, error)
	List(filter ListFilter, pagination common.PaginationQuery) ([]database.RoleGroup, int64, error)
	CodeExists(code string, excludeID int64) (bool, error)
	CreateWithPermissions(roleGroup *database.RoleGroup, permissions []database.RoleGroupPermission) error
	SaveWithPermissions(roleGroup *database.RoleGroup, permissions []database.RoleGroupPermission) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db.DB()}
}

func (r *repository) FindByID(id int64) (*database.RoleGroup, error) {
	var roleGroup database.RoleGroup
	if err := r.db.Preload("Permissions").First(&roleGroup, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleGroupNotFound
		}

		return nil, fmt.Errorf("find role group by id: %w", err)
	}

	return &roleGroup, nil
}

func (r *repository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.RoleGroup, int64, error) {
	query := r.db.Model(&database.RoleGroup{})

	if filter.Keyword != "" {
		keyword := "%" + strings.ToLower(filter.Keyword) + "%"
		query = query.Where(
			"LOWER(code) LIKE ? OR LOWER(name) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ?",
			keyword,
			keyword,
			keyword,
		)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count role groups: %w", err)
	}

	orderBy := buildOrderBy(pagination)

	var items []database.RoleGroup
	if err := query.Preload("Permissions").Order(orderBy).Offset(pagination.Offset()).Limit(pagination.PageSize).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list role groups: %w", err)
	}

	return items, total, nil
}

func (r *repository) CodeExists(code string, excludeID int64) (bool, error) {
	query := r.db.Model(&database.RoleGroup{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("count role groups by code: %w", err)
	}

	return count > 0, nil
}

func (r *repository) CreateWithPermissions(roleGroup *database.RoleGroup, permissions []database.RoleGroupPermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(roleGroup).Error; err != nil {
			return fmt.Errorf("create role group: %w", err)
		}

		if len(permissions) == 0 {
			return nil
		}

		for i := range permissions {
			permissions[i].RoleGroupID = roleGroup.ID
		}

		if err := tx.Create(&permissions).Error; err != nil {
			return fmt.Errorf("create role group permissions: %w", err)
		}

		roleGroup.Permissions = permissions
		return nil
	})
}

func (r *repository) SaveWithPermissions(roleGroup *database.RoleGroup, permissions []database.RoleGroupPermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(roleGroup).Error; err != nil {
			return fmt.Errorf("save role group: %w", err)
		}

		if err := tx.Where("role_group_id = ?", roleGroup.ID).Delete(&database.RoleGroupPermission{}).Error; err != nil {
			return fmt.Errorf("delete role group permissions: %w", err)
		}

		if len(permissions) == 0 {
			roleGroup.Permissions = nil
			return nil
		}

		for i := range permissions {
			permissions[i].RoleGroupID = roleGroup.ID
		}

		if err := tx.Create(&permissions).Error; err != nil {
			return fmt.Errorf("create role group permissions: %w", err)
		}

		roleGroup.Permissions = permissions
		return nil
	})
}

func buildOrderBy(pagination common.PaginationQuery) string {
	column := "created_at"
	switch pagination.SortBy {
	case "id", "code", "name", "created_at", "updated_at", "is_active":
		column = pagination.SortBy
	}

	return column + " " + pagination.SortOrder
}
