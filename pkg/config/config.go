package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Environment   string `mapstructure:"ENVIRONMENT"`
	Server        ServerConfig
	Database      DatabaseConfig
	Log           LogConfig   `mapstructure:"LOG"`
	Auth          AuthConfig  `mapstructure:"AUTH"`
	Cache         CacheConfig `mapstructure:"CACHE"`
	CORS          CORSConfig
	Secondary     []SecondaryConfig   `mapstructure:"SECONDARY"`
	Impersonation ImpersonationConfig `mapstructure:"IMPERSONATION"`
}

// ImpersonationConfig lets this service accept keeper-minted impersonation
// tokens (a sysadmin acting as a user). Tokens are signed with a dedicated
// secret and scoped to Audience; this service rejects any whose audience is not
// its own. When RevocationCheck is enabled, the auth middleware asks keeper
// whether a session is still active (cached for RevocationTTL) so sessions can
// be killed before their short expiry. Disabled by default.
type ImpersonationConfig struct {
	Enabled         bool          `mapstructure:"ENABLED"`
	JWTSecret       string        `mapstructure:"JWT_SECRET"`
	Audience        string        `mapstructure:"AUDIENCE"`
	KeeperBaseURL   string        `mapstructure:"KEEPER_BASE_URL"`
	RevocationCheck bool          `mapstructure:"REVOCATION_CHECK"`
	RevocationTTL   time.Duration `mapstructure:"REVOCATION_TTL"`
	RevocationHTTP  time.Duration `mapstructure:"REVOCATION_TIMEOUT"`
}

// SecondaryConfig drives one optional secondary listener: an additional HTTP
// server in the same process exposing only the allow-listed routes, with
// rate limiting configured independently of the primary server. Any number
// of listeners can be declared under SECONDARY. Identity always comes from
// JWT; JWT_SECRET (optional) makes the listener verify with a different
// signing key (e.g. keeper's guest secret) instead of AUTH.JWT_SECRET.
type SecondaryConfig struct {
	Name      string          `mapstructure:"NAME"`
	Enabled   bool            `mapstructure:"ENABLED"`
	Addr      string          `mapstructure:"ADDR"`
	JWTSecret string          `mapstructure:"JWT_SECRET"`
	RateLimit RateLimitConfig `mapstructure:"RATE_LIMIT"`
	Routes    []string        `mapstructure:"ROUTES"`
}

// RateLimitConfig holds rate limiter settings for a secondary listener.
type RateLimitConfig struct {
	Requests int           `mapstructure:"REQUESTS"`
	Window   time.Duration `mapstructure:"WINDOW"`
}

// CacheConfig holds in-memory cache configuration.
type CacheConfig struct {
	StatsTTL time.Duration `mapstructure:"STATS_TTL"`
}

// CORSConfig holds CORS-specific configuration.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`
}

// ServerConfig holds server related configuration.
type ServerConfig struct {
	Addr         string        `mapstructure:"ADDR"`
	Host         string        `mapstructure:"HOST"`
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"IDLE_TIMEOUT"`
}

// AuthConfig holds authentication related configuration.
type AuthConfig struct {
	JWTSecret string        `mapstructure:"JWT_SECRET"`
	JWTExpiry time.Duration `mapstructure:"JWT_EXPIRY"`
}

// DatabaseConfig holds database related configuration.
type DatabaseConfig struct {
	Driver string `mapstructure:"DRIVER"`
	Path   string `mapstructure:"PATH"`
	DSN    string `mapstructure:"DSN"`
}

// LogConfig holds logging-specific configuration.
type LogConfig struct {
	Dir   string `mapstructure:"DIR"`
	Level string `mapstructure:"LEVEL"`
}

// Load loads the configuration from files and environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// Default values
	v.SetDefault("ENVIRONMENT", "production")
	v.SetDefault("SERVER.ADDR", ":8081")
	v.SetDefault("SERVER.HOST", "localhost:8081")
	v.SetDefault("SERVER.READ_TIMEOUT", 5*time.Second)
	v.SetDefault("SERVER.WRITE_TIMEOUT", 10*time.Second)
	v.SetDefault("SERVER.IDLE_TIMEOUT", 120*time.Second)
	v.SetDefault("DATABASE.DRIVER", "sqlite3")
	v.SetDefault("DATABASE.PATH", "data/squirrel.db")
	v.SetDefault("DATABASE.DSN", "")
	v.SetDefault("LOG.DIR", "log")
	v.SetDefault("LOG.LEVEL", "info")
	v.SetDefault("AUTH.JWT_SECRET", "a-very-secure-and-shared-secret-key")
	v.SetDefault("AUTH.JWT_EXPIRY", 24*time.Hour)
	v.SetDefault("CACHE.STATS_TTL", 30*time.Second)
	v.SetDefault("IMPERSONATION.ENABLED", false)
	v.SetDefault("IMPERSONATION.JWT_SECRET", "a-separate-impersonation-token-secret-key")
	v.SetDefault("IMPERSONATION.AUDIENCE", "squirrel")
	v.SetDefault("IMPERSONATION.KEEPER_BASE_URL", "http://keeper:8080")
	v.SetDefault("IMPERSONATION.REVOCATION_CHECK", true)
	v.SetDefault("IMPERSONATION.REVOCATION_TTL", 30*time.Second)
	v.SetDefault("IMPERSONATION.REVOCATION_TIMEOUT", 5*time.Second)
	v.SetDefault("CORS.ALLOWED_ORIGINS", []string{"*"})

	// Environment variables
	v.SetEnvPrefix("SQUIRREL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// 1. Try to load base config.yaml
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read base config file: %w", err)
		}
	}

	// 2. Try to load environment-specific config (e.g. config.development.yaml)
	env := v.GetString("ENVIRONMENT")
	if env != "" {
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		if err := v.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to merge environment-specific config file: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := normalizeSecondary(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// normalizeSecondary validates the secondary listener entries and applies
// per-entry defaults (viper defaults cannot reach into list elements).
func normalizeSecondary(cfg *Config) error {
	seen := map[string]bool{cfg.Server.Addr: true}
	for i := range cfg.Secondary {
		s := &cfg.Secondary[i]
		if !s.Enabled {
			continue
		}
		if s.Name == "" {
			s.Name = fmt.Sprintf("secondary-%d", i)
		}
		if s.Addr == "" {
			return fmt.Errorf("SECONDARY[%d] (%s): ADDR is required", i, s.Name)
		}
		if seen[s.Addr] {
			return fmt.Errorf("SECONDARY[%d] (%s): ADDR %q already in use by another listener", i, s.Name, s.Addr)
		}
		seen[s.Addr] = true
		if s.RateLimit.Requests <= 0 {
			s.RateLimit.Requests = 100
		}
		if s.RateLimit.Window <= 0 {
			s.RateLimit.Window = 1 * time.Minute
		}
	}
	return nil
}
