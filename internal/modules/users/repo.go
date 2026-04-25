package users

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"school_enrollment_be/internal/common"
	"school_enrollment_be/internal/database"
)

type Repository interface {
	FindByID(id int64) (*database.User, error)
	List(filter ListFilter, pagination common.PaginationQuery) ([]database.User, int64, error)
	UsernameExists(username string, excludeID int64) (bool, error)
	EmailExists(email string, excludeID int64) (bool, error)
	Create(user *database.User) error
	Save(user *database.User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db.DB()}
}

func (r *repository) FindByID(id int64) (*database.User, error) {
	var user database.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return &user, nil
}

func (r *repository) List(filter ListFilter, pagination common.PaginationQuery) ([]database.User, int64, error) {
	query := r.db.Model(&database.User{})

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

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	orderBy := buildOrderBy(pagination)

	var items []database.User
	if err := query.Order(orderBy).Offset(pagination.Offset()).Limit(pagination.PageSize).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	return items, total, nil
}

func (r *repository) UsernameExists(username string, excludeID int64) (bool, error) {
	query := r.db.Model(&database.User{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("count users by username: %w", err)
	}

	return count > 0, nil
}

func (r *repository) EmailExists(email string, excludeID int64) (bool, error) {
	query := r.db.Model(&database.User{}).Where("email = ?", email)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("count users by email: %w", err)
	}

	return count > 0, nil
}

func (r *repository) Create(user *database.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *repository) Save(user *database.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("save user: %w", err)
	}

	return nil
}

func buildOrderBy(pagination common.PaginationQuery) string {
	column := "created_at"
	switch pagination.SortBy {
	case "id", "username", "full_name", "email", "created_at", "updated_at", "is_active":
		column = pagination.SortBy
	}

	return column + " " + pagination.SortOrder
}
