package config

import (
	"chatter/pkg/postgres"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres postgres.PostgresConfig `yaml:"postgres"`
	Server   ServerConfig            `yaml:"server"`
	Auth     AuthConfig              `yaml:"auth"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type AuthConfig struct {
	AccessTTL  time.Duration `yaml:"access_ttl"`
	RefreshTTL time.Duration `yaml:"refresh_ttl"`
	Secret     string        `yaml:"secret"`
}

func Load() *Config {
	var cfg Config
	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		return defaultConfig()
	}

	return &cfg
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{Addr: ":8080"},
		Auth:   AuthConfig{Secret: "a-string-secret-at-least-256-bits-long", AccessTTL: 24 * time.Hour, RefreshTTL: 1440 * time.Hour},
	}
}
