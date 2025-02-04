package config

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the configuration settings for the service.
type Config struct {
	Scope3 struct {
		APIURL string `mapstructure:"api_url"`
		Token  string `mapstructure:"token"`
	} `mapstructure:"scope3"`
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`
	Cache struct {
		DefaultTTL      string `mapstructure:"default_ttl"`
		CleanupInterval string `mapstructure:"cleanup_interval"`
	} `mapstructure:"cache"`
}

// LoadConfig reads configuration from the specified file, expanding environment variables.
func LoadConfig(path string) (*Config, error) {
	// Read the file content.
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Expand environment variables in the file content.
	expandedContent := os.ExpandEnv(string(b))

	// Set the config type and read from the expanded content.
	viper.SetConfigType("yaml")
	err = viper.ReadConfig(strings.NewReader(expandedContent))
	if err != nil {
		return nil, err
	}

	viper.AutomaticEnv() // Allow environment variables to override configuration values.

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetCacheTTL returns the TTL duration for cache entries.
func (c *Config) GetCacheTTL() (time.Duration, error) {
	return time.ParseDuration(c.Cache.DefaultTTL)
}

// GetCleanupInterval returns the interval at which the cache should be cleaned up.
func (c *Config) GetCleanupInterval() (time.Duration, error) {
	return time.ParseDuration(c.Cache.CleanupInterval)
}
