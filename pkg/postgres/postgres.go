package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Host     string `yaml:"host" env:"HOST"`
	Port     uint16 `yaml:"port" env:"PORT"`
	Username string `yaml:"username" env:"USER"`
	Password string `yaml:"password" env:"PASSWORD"`
	Database string `yaml:"database" env:"DB"`
	MinConns int32  `yaml:"min_conns" env:"MIN_CONNS"`
	MaxConns int32  `yaml:"max_conns" env:"MAX_CONNS"`
}

func New(ctx context.Context, config *PostgresConfig) (*pgxpool.Pool, error) {
	connString := config.connectionString()

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return conn, nil
}

func (c *PostgresConfig) connectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.MaxConns,
		c.MinConns,
	)
}
