package main

import (
	"fmt"
	"os"

	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/seed"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs"
	}
	cfg := config.LoadOrPanic(configPath)

	// Initialize logger
	if err := logger.Init("info"); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Get().Sync()

	log := logger.Get()
	log.Info("Starting database seeding...")

	// Initialize database
	db, err := database.InitGORM(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database", logger.Error(err))
	}
	defer db.Close()

	// Seed database
	if err := seed.SeedDatabase(db, cfg, log); err != nil {
		log.Fatal("Failed to seed database", logger.Error(err))
	}

	log.Info("Database seeding completed successfully")
	fmt.Println("✅ Database seeding completed successfully")
}
