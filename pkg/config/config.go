package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

// ServerConfig holds server related configuration.
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database related configuration.
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// LoggerConfig holds logger related configuration.
type LoggerConfig struct {
	Directory string `mapstructure:"directory"`
	Level     string `mapstructure:"level"`
}

// Load loads the configuration from files and environment variables.
func Load(env string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", 5*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.idle_timeout", 120*time.Second)
	v.SetDefault("database.path", "data/vyaya.db")
	v.SetDefault("logger.directory", "log")
	v.SetDefault("logger.level", "info")

	v.SetConfigName("config") // base config file name
	v.AddConfigPath("config") // path to look for the config file in
	v.SetConfigType("yaml")

	// Read base config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	// Read environment specific config
	if env != "" {
		v.SetConfigName("config." + env)
		if err := v.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("merge env config: %w", err)
			}
		}
	}

	// Read from environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind traditional environment variables
	_ = v.BindEnv("database.path", "DB_PATH")
	_ = v.BindEnv("logger.directory", "LOG_DIR")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
