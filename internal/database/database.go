package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jonosize/affiliate-platform/internal/config"
)

// DB holds both write and read database connections
type DB struct {
	Write *gorm.DB
	Read  *gorm.DB
}

// InitGORM initializes GORM without AutoMigrate
// Schema changes must be managed via SQL migrations (golang-migrate)
func InitGORM(cfg config.Config) (*DB, error) {
	writeURL := cfg.GetDatabaseWriteURL()
	readURL := cfg.GetDatabaseReadURL()

	// Write connection
	writeDB, err := gorm.Open(postgres.Open(writeURL), &gorm.Config{
		// Do NOT use AutoMigrate - schema managed by SQL migrations
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open write database: %w", err)
	}

	// Configure connection pool for write DB
	sqlWriteDB, err := writeDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get write database instance: %w", err)
	}

	sqlWriteDB.SetMaxOpenConns(cfg.GetDatabaseMaxOpenConns())
	sqlWriteDB.SetMaxIdleConns(cfg.GetDatabaseMaxIdleConns())
	sqlWriteDB.SetConnMaxLifetime(time.Duration(cfg.GetDatabaseConnMaxLifetime()) * time.Minute)

	// Test write connection
	if err := sqlWriteDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping write database: %w", err)
	}

	// Read connection (fallback to write if not configured)
	readURLToUse := readURL
	if readURLToUse == "" || readURLToUse == writeURL {
		readURLToUse = writeURL
	}

	readDB, err := gorm.Open(postgres.Open(readURLToUse), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open read database: %w", err)
	}

	// Configure connection pool for read DB
	sqlReadDB, err := readDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get read database instance: %w", err)
	}

	sqlReadDB.SetMaxOpenConns(cfg.GetDatabaseMaxOpenConns())
	sqlReadDB.SetMaxIdleConns(cfg.GetDatabaseMaxIdleConns())
	sqlReadDB.SetConnMaxLifetime(time.Duration(cfg.GetDatabaseConnMaxLifetime()) * time.Minute)

	// Test read connection
	if err := sqlReadDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping read database: %w", err)
	}

	return &DB{
		Write: writeDB,
		Read:  readDB,
	}, nil
}

// Close closes both database connections
func (db *DB) Close() error {
	sqlWriteDB, err := db.Write.DB()
	if err == nil {
		if err := sqlWriteDB.Close(); err != nil {
			return fmt.Errorf("failed to close write database: %w", err)
		}
	}

	sqlReadDB, err := db.Read.DB()
	if err == nil && sqlReadDB != sqlWriteDB {
		if err := sqlReadDB.Close(); err != nil {
			return fmt.Errorf("failed to close read database: %w", err)
		}
	}

	return nil
}
