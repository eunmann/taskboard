package app

import (
	"context"
	"fmt"

	"github.com/OWNER/PROJECT_NAME/internal/platform/db"
	"github.com/OWNER/PROJECT_NAME/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// Deps holds all shared application dependencies.
type Deps struct {
	Pool     *pgxpool.Pool
	Logger   zerolog.Logger
	UserRepo *user.Repo
}

// NewDeps creates a Deps container from configuration.
func NewDeps(ctx context.Context, cfg Config, logger zerolog.Logger) (*Deps, error) {
	pool, err := connectDatabase(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	return NewDepsWithPool(ctx, cfg, logger, pool), nil
}

// NewDepsWithPool creates a Deps container with an existing pool (for testing).
func NewDepsWithPool(_ context.Context, _ Config, logger zerolog.Logger, pool *pgxpool.Pool) *Deps {
	d := &Deps{
		Pool:   pool,
		Logger: logger,
	}

	d.UserRepo = user.NewRepo(pool)

	return d
}

// Close releases all resources held by dependencies.
func (d *Deps) Close() error {
	if d.Pool != nil {
		d.Pool.Close()
	}

	return nil
}

func connectDatabase(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dbCfg := db.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Database,
		SSLMode:  cfg.Database.SSLMode,
	}

	pool, err := db.Connect(ctx, dbCfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return pool, nil
}
