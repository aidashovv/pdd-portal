package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ConnectionPool struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewConnectionPool(ctx context.Context, cfg Config) (*ConnectionPool, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	pgxConfig.MaxConns = cfg.MaxConns
	pgxConfig.MinConns = cfg.MinConns

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.QueryTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pgx pool: %w", err)
	}

	return &ConnectionPool{
		pool:         pool,
		queryTimeout: cfg.QueryTimeout,
	}, nil
}

func (c *ConnectionPool) Pool() *pgxpool.Pool {
	return c.pool
}

func (c *ConnectionPool) QueryTimeout() time.Duration {
	return c.queryTimeout
}

func (c *ConnectionPool) Close() {
	if c == nil || c.pool == nil {
		return
	}

	c.pool.Close()
}
