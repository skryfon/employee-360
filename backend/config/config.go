package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// DefaultJWTSecret is the fallback secret used exclusively in local development mode.
const DefaultJWTSecret = "employee360-super-secret-key-change-in-production"

// Config represents the root application configuration for Employee360.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	App      AppConfig      `mapstructure:"app"`
}

// ServerConfig contains HTTP server configuration parameters.
type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Address returns the formatted host:port listener address.
func (s ServerConfig) Address() string {
	if s.Host == "" {
		return ":" + s.Port
	}
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

// DatabaseConfig contains PostgreSQL connection and pooling parameters.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnTimeout     time.Duration `mapstructure:"conn_timeout"`
}

// DSN returns the PostgreSQL DSN formatted for GORM pgx/postgres driver.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// URL returns the PostgreSQL connection URL formatted for drivers and migrators.
func (d DatabaseConfig) URL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

// AppConfig contains generic application metadata and logging settings.
type AppConfig struct {
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
}

// Validate verifies that the configuration meets environment and security requirements.
func (c *Config) Validate() error {
	env := strings.ToLower(c.App.Environment)
	if env == "production" || env == "staging" {
		if c.JWT.Secret == DefaultJWTSecret || len(c.JWT.Secret) < 32 {
			return fmt.Errorf("jwt.secret must be explicitly set to a custom strong secret (min 32 chars) in %q environment", c.App.Environment)
		}
	}
	return nil
}

// Load loads configuration from defaults, .env files, config.yaml files, and environment variables.
func Load(searchPaths ...string) (*Config, error) {
	v := viper.New()

	// 1. Set Defaults
	setDefaults(v)

	// 2. Load .env if present (look in current dir, parents, and searchPaths)
	loadDotEnv(searchPaths...)

	// 3. Configure Viper for config.yaml / config.yml
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Standard config search paths
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")
	v.AddConfigPath("../../config")

	for _, path := range searchPaths {
		if path != "" {
			v.AddConfigPath(path)
		}
	}

	// 4. Environment variable handling
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit environment variable bindings & aliases
	bindEnvAliases(v)

	// 5. Read config file if it exists (ignore ErrFileNotFound)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// If file was found but parsing failed, return error
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// 6. Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode configuration into struct: %w", err)
	}

	// 7. Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.shutdown_timeout", 10*time.Second)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.dbname", "employee360")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 15*time.Minute)
	v.SetDefault("database.conn_timeout", 5*time.Second)

	// JWT defaults
	v.SetDefault("jwt.secret", DefaultJWTSecret)
	v.SetDefault("jwt.access_expiry", 15*time.Minute)
	v.SetDefault("jwt.refresh_expiry", 7*24*time.Hour)

	// App defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.log_level", "debug")
}

func loadDotEnv(searchPaths ...string) {
	candidates := []string{
		".env",
		"backend/.env",
		"../.env",
		"../../.env",
	}

	for _, sp := range searchPaths {
		if sp != "" {
			candidates = append(candidates, filepath.Join(sp, ".env"))
		}
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Overload(path)
			break
		}
	}
}

func bindEnvAliases(v *viper.Viper) {
	_ = v.BindEnv("server.port", "SERVER_PORT", "PORT")
	_ = v.BindEnv("server.host", "SERVER_HOST", "HOST")
	_ = v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	_ = v.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")

	_ = v.BindEnv("database.host", "DATABASE_HOST", "DB_HOST")
	_ = v.BindEnv("database.port", "DATABASE_PORT", "DB_PORT")
	_ = v.BindEnv("database.user", "DATABASE_USER", "DB_USER", "POSTGRES_USER")
	_ = v.BindEnv("database.password", "DATABASE_PASSWORD", "DB_PASSWORD", "POSTGRES_PASSWORD")
	_ = v.BindEnv("database.dbname", "DATABASE_NAME", "DATABASE_DBNAME", "DB_NAME", "POSTGRES_DB")
	_ = v.BindEnv("database.sslmode", "DATABASE_SSLMODE", "DB_SSLMODE")
	_ = v.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS", "DB_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS", "DB_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME", "DB_CONN_MAX_LIFETIME")
	_ = v.BindEnv("database.conn_timeout", "DATABASE_CONN_TIMEOUT", "DB_CONN_TIMEOUT")

	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("jwt.access_expiry", "JWT_ACCESS_EXPIRY")
	_ = v.BindEnv("jwt.refresh_expiry", "JWT_REFRESH_EXPIRY")

	_ = v.BindEnv("app.environment", "APP_ENV", "ENVIRONMENT", "ENV")
	_ = v.BindEnv("app.log_level", "LOG_LEVEL", "APP_LOG_LEVEL")
}
