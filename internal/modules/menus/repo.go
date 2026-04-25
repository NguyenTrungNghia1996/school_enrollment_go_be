package menus

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
)

type Repository interface {
	FindByID(id int64) (*database.Menu, error)
	List(filter ListFilter, pagination common.PaginationQuery) ([]database.Menu, int64, error)
	KeyCodeExists(keyCode string, excludeID int64) (bool, error)
	Create(menu *database.Menu) error
	Save(menu *database.Menu) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db.DB()}
}

func (r *repository) FindByID(id int64) (*database.Menu, error) {
	var menu database.Menu
	if err := r.db.First(&menu, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuNotFound
		}

		return nil, fmt.Errorf("find menu by id: %w", err)
	}

	return &menu, nil
}

func (r *repository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.Menu, int64, error) {
	query := r.db.Model(&database.Menu{})

	if filter.Keyword != "" {
		keyword := "%" + strings.ToLower(filter.Keyword) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(key_code) LIKE ? OR LOWER(COALESCE(url, '')) LIKE ?",
			keyword,
			keyword,
			keyword,
		)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.ParentID != nil {
		query = query.Where("parent_id = ?", *filter.ParentID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count menus: %w", err)
	}

	orderBy := buildOrderBy(pagination)
	query = query.Order(orderBy)

	if filter.Paginate {
		query = query.Offset(pagination.Offset()).Limit(pagination.PageSize)
	}

	var items []database.Menu
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list menus: %w", err)
	}

	return items, total, nil
}

func (r *repository) KeyCodeExists(keyCode string, excludeID int64) (bool, error) {
	query := r.db.Model(&database.Menu{}).Where("key_code = ?", keyCode)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("count menus by key code: %w", err)
	}

	return count > 0, nil
}

func (r *repository) Create(menu *database.Menu) error {
	if err := r.db.Create(menu).Error; err != nil {
		return fmt.Errorf("create menu: %w", err)
	}

	return nil
}

func (r *repository) Save(menu *database.Menu) error {
	if err := r.db.Save(menu).Error; err != nil {
		return fmt.Errorf("save menu: %w", err)
	}

	return nil
}

func buildOrderBy(pagination common.PaginationQuery) string {
	column := "sort_order"
	switch pagination.SortBy {
	case "id", "parent_id", "title", "key_code", "permission_bit", "is_active", "sort_order", "created_at", "updated_at":
		column = pagination.SortBy
	}

	return column + " " + pagination.SortOrder
}
