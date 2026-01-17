package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type viperConfig struct {
	v *viper.Viper
}

// NewViperConfig creates a new config instance using viper
func NewViperConfig(configPath string) (Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file path and type
	v.SetConfigName("config")
	v.SetConfigType("json")
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath(".")         // Look in current directory
	v.AddConfigPath("./configs") // Look in configs directory

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is OK if using env vars or .env
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Enable .env file support
	v.SetConfigType("env")
	v.SetConfigName(".env")
	v.AddConfigPath(".")
	// Try to read .env file (optional - not error if not found)
	if err := v.MergeInConfig(); err != nil {
		// .env file not found is OK
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading .env file: %w", err)
		}
	}

	// Enable environment variables (env vars override .env and config.json)
	// Priority: env vars > .env file > config.json > defaults
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return &viperConfig{v: v}, nil
}

func setDefaults(v *viper.Viper) {
	// Database Write defaults
	v.SetDefault("database.write.host", "localhost")
	v.SetDefault("database.write.port", 5432)
	v.SetDefault("database.write.user", "jonosize")
	v.SetDefault("database.write.password", "jonosize_dev")
	v.SetDefault("database.write.dbname", "jonosize")
	v.SetDefault("database.write.sslmode", "disable")

	// Database Read: No defaults - will fallback to write values if not set
	// Only set defaults if you want to override specific read parameters

	// Database Connection Pool defaults
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5) // minutes

	// Redis defaults
	v.SetDefault("redis.url", "")

	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.host", "0.0.0.0")

	// API defaults
	v.SetDefault("api.base_url", "http://localhost:8080")

	// Worker defaults (6-field format: second minute hour day month weekday)
	v.SetDefault("worker.price_refresh_cron", "0 0 */6 * * *")

	// Adapters defaults
	v.SetDefault("adapters.mock_mode", false)

	// Auth defaults (empty - must be provided via env or config)
	v.SetDefault("auth.basic_auth.username", "")
	v.SetDefault("auth.basic_auth.password", "")
}

// Implement Config interface - Database Write
func (c *viperConfig) GetDatabaseWriteHost() string {
	return c.v.GetString("database.write.host")
}

func (c *viperConfig) GetDatabaseWritePort() int {
	return c.v.GetInt("database.write.port")
}

func (c *viperConfig) GetDatabaseWriteUser() string {
	return c.v.GetString("database.write.user")
}

func (c *viperConfig) GetDatabaseWritePassword() string {
	return c.v.GetString("database.write.password")
}

func (c *viperConfig) GetDatabaseWriteDBName() string {
	return c.v.GetString("database.write.dbname")
}

func (c *viperConfig) GetDatabaseWriteSSLMode() string {
	return c.v.GetString("database.write.sslmode")
}

// Implement Config interface - Database Read (always fallback to write if not set)
func (c *viperConfig) GetDatabaseReadHost() string {
	// If read.host is explicitly set and not empty, use it
	if c.v.IsSet("database.read.host") {
		if host := c.v.GetString("database.read.host"); host != "" {
			return host
		}
	}
	// Otherwise, use write host
	return c.GetDatabaseWriteHost()
}

func (c *viperConfig) GetDatabaseReadPort() int {
	// If read.port is explicitly set and not zero, use it
	if c.v.IsSet("database.read.port") {
		if port := c.v.GetInt("database.read.port"); port != 0 {
			return port
		}
	}
	// Otherwise, use write port
	return c.GetDatabaseWritePort()
}

func (c *viperConfig) GetDatabaseReadUser() string {
	// If read.user is explicitly set and not empty, use it
	if c.v.IsSet("database.read.user") {
		if user := c.v.GetString("database.read.user"); user != "" {
			return user
		}
	}
	// Otherwise, use write user
	return c.GetDatabaseWriteUser()
}

func (c *viperConfig) GetDatabaseReadPassword() string {
	// If read.password is explicitly set and not empty, use it
	if c.v.IsSet("database.read.password") {
		if password := c.v.GetString("database.read.password"); password != "" {
			return password
		}
	}
	// Otherwise, use write password
	return c.GetDatabaseWritePassword()
}

func (c *viperConfig) GetDatabaseReadDBName() string {
	// If read.dbname is explicitly set and not empty, use it
	if c.v.IsSet("database.read.dbname") {
		if dbname := c.v.GetString("database.read.dbname"); dbname != "" {
			return dbname
		}
	}
	// Otherwise, use write dbname
	return c.GetDatabaseWriteDBName()
}

func (c *viperConfig) GetDatabaseReadSSLMode() string {
	// If read.sslmode is explicitly set and not empty, use it
	if c.v.IsSet("database.read.sslmode") {
		if sslmode := c.v.GetString("database.read.sslmode"); sslmode != "" {
			return sslmode
		}
	}
	// Otherwise, use write sslmode
	return c.GetDatabaseWriteSSLMode()
}

// Helper: Build connection URL from parameters
func (c *viperConfig) GetDatabaseWriteURL() string {
	return buildPostgresURL(
		c.GetDatabaseWriteHost(),
		c.GetDatabaseWritePort(),
		c.GetDatabaseWriteUser(),
		c.GetDatabaseWritePassword(),
		c.GetDatabaseWriteDBName(),
		c.GetDatabaseWriteSSLMode(),
	)
}

func (c *viperConfig) GetDatabaseReadURL() string {
	return buildPostgresURL(
		c.GetDatabaseReadHost(),
		c.GetDatabaseReadPort(),
		c.GetDatabaseReadUser(),
		c.GetDatabaseReadPassword(),
		c.GetDatabaseReadDBName(),
		c.GetDatabaseReadSSLMode(),
	)
}

// Helper function to build PostgreSQL connection URL
func buildPostgresURL(host string, port int, user, password, dbname, sslmode string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

// Database Connection Pool
func (c *viperConfig) GetDatabaseMaxOpenConns() int {
	return c.v.GetInt("database.max_open_conns")
}

func (c *viperConfig) GetDatabaseMaxIdleConns() int {
	return c.v.GetInt("database.max_idle_conns")
}

func (c *viperConfig) GetDatabaseConnMaxLifetime() int {
	return c.v.GetInt("database.conn_max_lifetime")
}

func (c *viperConfig) GetRedisURL() string {
	return c.v.GetString("redis.url")
}

func (c *viperConfig) GetServerPort() string {
	return c.v.GetString("server.port")
}

func (c *viperConfig) GetServerHost() string {
	return c.v.GetString("server.host")
}

func (c *viperConfig) GetAPIBaseURL() string {
	return c.v.GetString("api.base_url")
}

func (c *viperConfig) GetPriceRefreshCron() string {
	return c.v.GetString("worker.price_refresh_cron")
}

func (c *viperConfig) GetMockMode() bool {
	return c.v.GetBool("adapters.mock_mode")
}

func (c *viperConfig) GetBasicAuthUsername() string {
	return c.v.GetString("auth.basic_auth.username")
}

func (c *viperConfig) GetBasicAuthPassword() string {
	return c.v.GetString("auth.basic_auth.password")
}

func (c *viperConfig) GetAllSettings() map[string]interface{} {
	return c.v.AllSettings()
}
