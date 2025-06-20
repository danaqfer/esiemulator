package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the emulator suite
type Config struct {
	// Server configuration
	Port int
	Host string

	// Emulator configuration
	EmulatorMode string
	ESIMode      string
	Debug        bool

	// Logging configuration
	LogLevel string
	LogFile  string

	// Performance configuration
	MaxConcurrentRequests int
	RequestTimeout        int

	// Cache configuration
	CacheEnabled bool
	CacheSize    int
	CacheTTL     int
}

// Default configuration values
const (
	DefaultPort                  = 3000
	DefaultHost                  = "localhost"
	DefaultEmulatorMode          = "esi"
	DefaultESIMode               = "akamai"
	DefaultLogLevel              = "info"
	DefaultMaxConcurrentRequests = 1000
	DefaultRequestTimeout        = 30
	DefaultCacheSize             = 1000
	DefaultCacheTTL              = 3600
)

// Load loads configuration from environment variables and defaults
func Load() *Config {
	config := &Config{
		Port:                  getEnvAsInt("PORT", DefaultPort),
		Host:                  getEnvAsString("HOST", DefaultHost),
		EmulatorMode:          getEnvAsString("EMULATOR_MODE", DefaultEmulatorMode),
		ESIMode:               getEnvAsString("ESI_MODE", DefaultESIMode),
		Debug:                 getEnvAsBool("DEBUG", false),
		LogLevel:              getEnvAsString("LOG_LEVEL", DefaultLogLevel),
		LogFile:               getEnvAsString("LOG_FILE", ""),
		MaxConcurrentRequests: getEnvAsInt("MAX_CONCURRENT_REQUESTS", DefaultMaxConcurrentRequests),
		RequestTimeout:        getEnvAsInt("REQUEST_TIMEOUT", DefaultRequestTimeout),
		CacheEnabled:          getEnvAsBool("CACHE_ENABLED", true),
		CacheSize:             getEnvAsInt("CACHE_SIZE", DefaultCacheSize),
		CacheTTL:              getEnvAsInt("CACHE_TTL", DefaultCacheTTL),
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate emulator mode
	validEmulatorModes := []string{"esi", "property-manager", "integrated"}
	if !contains(validEmulatorModes, c.EmulatorMode) {
		return &ConfigError{
			Field:   "EMULATOR_MODE",
			Value:   c.EmulatorMode,
			Message: "must be one of: " + strings.Join(validEmulatorModes, ", "),
		}
	}

	// Validate ESI mode
	validESIModes := []string{"fastly", "akamai", "w3c", "development"}
	if !contains(validESIModes, c.ESIMode) {
		return &ConfigError{
			Field:   "ESI_MODE",
			Value:   c.ESIMode,
			Message: "must be one of: " + strings.Join(validESIModes, ", "),
		}
	}

	// Validate port
	if c.Port < 1 || c.Port > 65535 {
		return &ConfigError{
			Field:   "PORT",
			Value:   strconv.Itoa(c.Port),
			Message: "must be between 1 and 65535",
		}
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return &ConfigError{
			Field:   "LOG_LEVEL",
			Value:   c.LogLevel,
			Message: "must be one of: " + strings.Join(validLogLevels, ", "),
		}
	}

	return nil
}

// GetAddress returns the full address string for the server
func (c *Config) GetAddress() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

// IsESIMode returns true if the emulator is in ESI mode
func (c *Config) IsESIMode() bool {
	return c.EmulatorMode == "esi"
}

// IsPropertyManagerMode returns true if the emulator is in Property Manager mode
func (c *Config) IsPropertyManagerMode() bool {
	return c.EmulatorMode == "property-manager"
}

// IsIntegratedMode returns true if the emulator is in integrated mode
func (c *Config) IsIntegratedMode() bool {
	return c.EmulatorMode == "integrated"
}

// IsDebugMode returns true if debug mode is enabled
func (c *Config) IsDebugMode() bool {
	return c.Debug
}

// Helper functions for environment variable parsing
func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Value   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error: " + e.Field + "='" + e.Value + "' " + e.Message
}
