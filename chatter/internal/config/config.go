package config

import (
	"chatter/pkg/postgres"
	"chatter/pkg/redis"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres postgres.PostgresConfig `yaml:"postgres" env-prefix:"POSTGRES_"`
	Redis    redis.RedisConfig       `yaml:"redis" env-prefix:"REDIS_"`
	Server   ServerConfig            `yaml:"server" env-prefix:"SERVER_"`
	Auth     AuthConfig              `yaml:"auth" env-prefix:"AUTH_"`
}

type ServerConfig struct {
	Addr        string   `yaml:"addr" env:"ADDR"`
	CorsOrigins []string `yaml:"cors_origins" env:"CORS_ORIGINS" env-separator:","`
}

type AuthConfig struct {
	AccessTTL  time.Duration `yaml:"access_ttl" env:"ACCESS_TTL"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env:"REFRESH_TTL"`
	Secret     string        `yaml:"secret" env:"SECRET"`
}

func Load() *Config {
	var cfg Config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return defaultConfig()
		}
	}

	return &cfg
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:        ":8080",
			CorsOrigins: []string{"http://localhost:5173"},
		},
		Auth: AuthConfig{
			Secret:     "a-string-secret-at-least-256-bits-long",
			AccessTTL:  24 * time.Hour,
			RefreshTTL: 1440 * time.Hour,
		},
	}
}
