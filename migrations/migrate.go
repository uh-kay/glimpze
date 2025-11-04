package migrations

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrationsFS embed.FS

func RunMigrations(pool *pgxpool.Pool) error {
	goose.SetBaseFS(migrationsFS)

	err := goose.SetDialect(string(goose.DialectPostgres))
	if err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pool)
	if err = goose.UpContext(context.Background(), db, "."); err != nil {
		return err
	}

	return nil
}
