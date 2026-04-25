package adminauth

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"school_enrollment_be/internal/database"
)

type Repository interface {
	FindByUsername(username string) (*database.AdminUser, error)
	FindByID(id int64) (*database.AdminUser, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db.DB()}
}

func (r *repository) FindByUsername(username string) (*database.AdminUser, error) {
	var admin database.AdminUser
	if err := r.db.Where("username = ?", username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("find admin by username: %w", err)
	}

	return &admin, nil
}

func (r *repository) FindByID(id int64) (*database.AdminUser, error) {
	var admin database.AdminUser
	if err := r.db.Preload("RoleGroups.Permissions").First(&admin, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}

		return nil, fmt.Errorf("find admin by id: %w", err)
	}

	return &admin, nil
}
