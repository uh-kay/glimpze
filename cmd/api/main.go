package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/uh-kay/glimpze/db"
	"github.com/uh-kay/glimpze/env"
	"github.com/uh-kay/glimpze/migrations"
	"github.com/uh-kay/glimpze/store"
	"github.com/uh-kay/glimpze/store/cache"
	"github.com/valkey-io/valkey-go"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":3000"),
		env:  env.GetString("ENV", "dev"),
		dbConfig: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://root:password@localhost:5432/glimpze?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		valkeyCfg: valkeyCfg{
			enabled: env.GetBool("VALKEY_ENABLED", false),
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := db.New(
		cfg.dbConfig.addr,
		cfg.dbConfig.maxOpenConns,
		cfg.dbConfig.maxIdleTime,
	)
	if err != nil {
		logger.Error("error connecting to db", "error", err.Error())
		log.Fatal(err)
	}

	store := store.NewStorage(db)

	var vdb valkey.Client
	if cfg.valkeyCfg.enabled {
		vdb, err = cache.NewValkeyClient(cfg.valkeyCfg.addr, cfg.valkeyCfg.pw, cfg.valkeyCfg.db)
		if err != nil {
			logger.Error("error connecting to valkey", "error", err.Error())
			log.Fatal(err)
		}
		defer vdb.Close()
	}

	cache := cache.NewValkeyStorage(vdb)

	// Run migrations
	if err := migrations.RunMigrations(db); err != nil {
		logger.Error("error migrating", "error", err.Error())
	}

	app := &application{
		logger: logger,
		config: cfg,
		db:     db,
		store:  store,
		cache:  cache,
	}

	err = app.run(app.mount())
	if err != nil {
		logger.Error("error starting server", "error", err.Error())
		log.Fatal(err)
	}
}
