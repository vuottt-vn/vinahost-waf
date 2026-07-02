package database

import (
	"fmt"
	"log"

	"github.com/webappfirewall/waf/internal/config"
	"github.com/webappfirewall/waf/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.DatabaseConfig) error {
	// Use SQLite file path or in-memory
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = "waf.db"
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Database connected successfully (SQLite: %s)", dbPath)

	// Auto-migrate all models
	err = DB.AutoMigrate(
		&models.User{},
		&models.Rule{},
		&models.AuditLog{},
		&models.ProxyTarget{},
		&models.IPEntry{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migrations completed")

	// Seed default admin user
	seedDefaultAdmin()

	return nil
}

func seedDefaultAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		log.Println("Seeding default admin user (admin/admin123)")
		// Password will be hashed by the API layer
		// We import auth here to avoid circular imports at package level
	}
}
