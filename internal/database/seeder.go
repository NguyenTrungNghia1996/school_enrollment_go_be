package database

import (
	"fmt"

	"gorm.io/gorm"

	"school_enrollment_be/internal/common/security"
)

const defaultPasswordHashCost = 0

func (d *Database) SeedDefaults() error {
	if !d.cfg.EnableDefaultSeed {
		return nil
	}

	return d.gorm.Transaction(func(tx *gorm.DB) error {
		hasher := security.NewPasswordHasher(defaultPasswordHashCost)

		if err := seedDefaultAdminUser(tx, hasher); err != nil {
			return err
		}

		if err := seedDefaultUser(tx, hasher); err != nil {
			return err
		}

		return nil
	})
}

func seedDefaultAdminUser(tx *gorm.DB, hasher security.PasswordHasher) error {
	var count int64
	if err := tx.Model(&AdminUser{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count admin users: %w", err)
	}

	if count > 0 {
		return nil
	}

	passwordHash, err := hasher.Hash("admin")
	if err != nil {
		return fmt.Errorf("hash default admin password: %w", err)
	}

	adminUser := AdminUser{
		Username:     "admin",
		PasswordHash: passwordHash,
		FullName:     "Default Admin",
		IsSuperAdmin: true,
		IsActive:     true,
	}

	if err := tx.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("create default admin user: %w", err)
	}

	return nil
}

func seedDefaultUser(tx *gorm.DB, hasher security.PasswordHasher) error {
	var count int64
	if err := tx.Model(&User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count users: %w", err)
	}

	if count > 0 {
		return nil
	}

	passwordHash, err := hasher.Hash("user")
	if err != nil {
		return fmt.Errorf("hash default user password: %w", err)
	}

	user := User{
		Username:     "user",
		PasswordHash: passwordHash,
		FullName:     "Default User",
		IsActive:     true,
	}

	if err := tx.Create(&user).Error; err != nil {
		return fmt.Errorf("create default user: %w", err)
	}

	return nil
}
