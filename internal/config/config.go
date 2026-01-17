package config

// Config interface provides abstraction layer
type Config interface {
	// Database Write
	GetDatabaseWriteHost() string
	GetDatabaseWritePort() int
	GetDatabaseWriteUser() string
	GetDatabaseWritePassword() string
	GetDatabaseWriteDBName() string
	GetDatabaseWriteSSLMode() string

	// Database Read
	GetDatabaseReadHost() string
	GetDatabaseReadPort() int
	GetDatabaseReadUser() string
	GetDatabaseReadPassword() string
	GetDatabaseReadDBName() string
	GetDatabaseReadSSLMode() string

	// Database Connection Pool
	GetDatabaseMaxOpenConns() int
	GetDatabaseMaxIdleConns() int
	GetDatabaseConnMaxLifetime() int // minutes

	// Helper: Build connection URL from parameters
	GetDatabaseWriteURL() string
	GetDatabaseReadURL() string

	// Redis
	GetRedisURL() string

	// Server
	GetServerPort() string
	GetServerHost() string

	// API
	GetAPIBaseURL() string

	// Worker
	GetPriceRefreshCron() string

	// Adapters
	GetMockMode() bool

	// Authentication (Basic Auth)
	GetBasicAuthUsername() string
	GetBasicAuthPassword() string

	// All settings
	GetAllSettings() map[string]interface{}
}

// DatabaseConfig holds database connection parameters
type DatabaseConfig struct {
	Host     string `json:"host,omitempty" mapstructure:"host"`
	Port     int    `json:"port,omitempty" mapstructure:"port"`
	User     string `json:"user,omitempty" mapstructure:"user"`
	Password string `json:"password,omitempty" mapstructure:"password"`
	DBName   string `json:"dbname,omitempty" mapstructure:"dbname"`
	SSLMode  string `json:"sslmode,omitempty" mapstructure:"sslmode"`
}

// AppConfig struct holds all configuration values
type AppConfig struct {
	Database struct {
		Write           DatabaseConfig  `json:"write" mapstructure:"write"`
		Read            *DatabaseConfig `json:"read,omitempty" mapstructure:"read"` // Optional: if not set, uses write values
		MaxOpenConns    int             `json:"max_open_conns" mapstructure:"max_open_conns"`
		MaxIdleConns    int             `json:"max_idle_conns" mapstructure:"max_idle_conns"`
		ConnMaxLifetime int             `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"` // minutes
	} `json:"database" mapstructure:"database"`

	Redis struct {
		URL string `json:"url" mapstructure:"url"`
	} `json:"redis" mapstructure:"redis"`

	Server struct {
		Port string `json:"port" mapstructure:"port"`
		Host string `json:"host" mapstructure:"host"`
	} `json:"server" mapstructure:"server"`

	API struct {
		BaseURL string `json:"base_url" mapstructure:"base_url"`
	} `json:"api" mapstructure:"api"`

	Worker struct {
		PriceRefreshCron string `json:"price_refresh_cron" mapstructure:"price_refresh_cron"`
	} `json:"worker" mapstructure:"worker"`

	Adapters struct {
		MockMode bool `json:"mock_mode" mapstructure:"mock_mode"`
	} `json:"adapters" mapstructure:"adapters"`

	Auth struct {
		BasicAuth struct {
			Username string `json:"username" mapstructure:"username"`
			Password string `json:"password" mapstructure:"password"`
		} `json:"basic_auth" mapstructure:"basic_auth"`
	} `json:"auth" mapstructure:"auth"`
}
