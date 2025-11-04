package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(addr string, maxOpenConns int, maxIdleTime string) (*pgxpool.Pool, error) {
	idleDuration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(addr)
	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(maxOpenConns)
	config.MaxConnIdleTime = idleDuration

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
