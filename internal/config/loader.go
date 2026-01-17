package config

import (
	"fmt"
)

var globalConfig Config

// Load initializes and loads configuration
// Priority: env vars > .env file > config.json > defaults
// Required values (if missing and no default): panic
func Load(configPath string) error {
	cfg, err := NewViperConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate required values (panic if missing)
	validateRequired(cfg)

	globalConfig = cfg
	return nil
}

// validateRequired checks for required config values and panics if missing
func validateRequired(cfg Config) {
	// Example: If DB password is required and empty, panic
	// Customize based on your requirements
	dbPassword := cfg.GetDatabaseWritePassword()
	if dbPassword == "" {
		panic("required config: database.write.password is empty")
	}

	// Add other required validations as needed
}

// Get returns the global config instance
func Get() Config {
	if globalConfig == nil {
		panic("config not loaded. Call config.Load() first")
	}
	return globalConfig
}

// LoadOrPanic loads config or panics on error
func LoadOrPanic(configPath string) Config {
	if err := Load(configPath); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return Get()
}
