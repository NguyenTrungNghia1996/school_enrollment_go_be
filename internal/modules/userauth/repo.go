package userauth

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"school_enrollment_be/internal/database"
)

type Repository interface {
	FindByUsername(username string) (*database.User, error)
	FindByID(id int64) (*database.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *database.Database) Repository {
	return &repository{db: db.DB()}
}

func (r *repository) FindByUsername(username string) (*database.User, error) {
	var user database.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return &user, nil
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
