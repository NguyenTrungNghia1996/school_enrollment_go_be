package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"school_enrollment_be/internal/config"
)

type Database struct {
	cfg   config.DatabaseConfig
	gorm  *gorm.DB
	sqlDB *sql.DB
}

func New(cfg *config.Config) (*Database, error) {
	db, err := gorm.Open(postgres.Open(buildDSN(cfg)), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: false,
		Logger:                                   newLogger(cfg.App.Env),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	if err := registerAssociations(db); err != nil {
		return nil, fmt.Errorf("register database associations: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db from gorm: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	database := &Database{
		cfg:   cfg.Database,
		gorm:  db,
		sqlDB: sqlDB,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.HealthCheck(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return database, nil
}

func registerAssociations(db *gorm.DB) error {
	if err := db.SetupJoinTable(&AdminUser{}, "RoleGroups", &AdminUserRoleGroup{}); err != nil {
		return err
	}

	if err := db.SetupJoinTable(&RoleGroup{}, "AdminUsers", &AdminUserRoleGroup{}); err != nil {
		return err
	}

	return nil
}

func (d *Database) DB() *gorm.DB {
	return d.gorm
}

func (d *Database) SQLDB() *sql.DB {
	return d.sqlDB
}

func (d *Database) HealthCheck(ctx context.Context) error {
	return d.sqlDB.PingContext(ctx)
}

func (d *Database) AutoMigrate(models ...interface{}) error {
	if !d.cfg.EnableAutoMigrate || len(models) == 0 {
		return nil
	}

	if err := d.gorm.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}

func (d *Database) Close() error {
	if d.sqlDB == nil {
		return nil
	}

	return d.sqlDB.Close()
}

func buildDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
		cfg.App.Timezone,
	)
}

func newLogger(appEnv string) gormlogger.Interface {
	logLevel := gormlogger.Warn
	if isDevEnv(appEnv) {
		logLevel = gormlogger.Info
	}

	return gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

func isDevEnv(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "local", "dev", "development":
		return true
	default:
		return false
	}
}
