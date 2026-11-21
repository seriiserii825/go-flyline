package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPPort     string        `envconfig:"HTTP_PORT" required:"true"`
	GinMode      string        `envconfig:"GIN_MODE" default:"debug"`
	DatabaseURL  string        `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret    string        `envconfig:"JWT_SECRET" required:"true"`
	JWTAccessTTL time.Duration `envconfig:"JWT_ACCESS_TTL" required:"true"`
	RefreshTTL   time.Duration `envconfig:"JWT_REFRESH_TTL" required:"true"`
}

func Load() (Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// HTTPAddr — адрес для r.Run / http.Server, например ":8080".
func (c Config) HTTPAddr() string {
	return ":" + c.HTTPPort
}

func (c Config) validate() error {
	switch c.GinMode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("GIN_MODE must be debug, release or test, got %q", c.GinMode)
	}

	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}

	if c.JWTAccessTTL <= 0 || c.RefreshTTL <= 0 {
		return errors.New("JWT_ACCESS_TTL and REFRESH_TTL must be positive")
	}

	if c.RefreshTTL <= c.JWTAccessTTL {
		return errors.New("REFRESH_TTL must be longer than JWT_ACCESS_TTL")
	}

	return nil
}
