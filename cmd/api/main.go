package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/uh-kay/glimpze/db"
	"github.com/uh-kay/glimpze/env"
	"github.com/uh-kay/glimpze/migrations"
	"github.com/uh-kay/glimpze/storage"
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
		r2Cfg: r2Cfg{
			bucketName:      env.GetString("R2_BUCKET_NAME", ""),
			accountID:       env.GetString("R2_ACCOUNT_ID", ""),
			accessKeyID:     env.GetString("R2_ACCESS_KEY_ID", ""),
			accessKeySecret: env.GetString("R2_ACCESS_KEY_SECRET", ""),
		},
		rateLimitCfg: rateLimitCfg{
			requestCount: env.GetInt("RATE_LIMITER_REQUEST_COUNT", 20),
			windowLength: time.Minute,
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

	// Valkey
	cache := cache.NewValkeyStorage(vdb)

	// Cloudflare R2
	storage, err := storage.NewR2Client(
		context.Background(),
		cfg.r2Cfg.bucketName,
		cfg.r2Cfg.accountID,
		cfg.r2Cfg.accessKeyID,
		cfg.r2Cfg.accessKeySecret,
	)
	if err != nil {
		logger.Error("error connecting to r2", "error", err.Error())
		log.Fatal(err)
	}

	// Run migrations
	if err := migrations.RunMigrations(db); err != nil {
		logger.Error("error migrating", "error", err.Error())
		log.Fatal(err)
	}

	// Cache default role
	defaultRole, err := store.Roles.GetByName(context.Background(), "user")
	if err != nil {
		logger.Error("error getting user role", "error", err.Error())
		log.Fatal(err)
	}

	app := &application{
		logger:      logger,
		config:      cfg,
		db:          db,
		store:       store,
		cache:       cache,
		storage:     storage,
		defaultRole: defaultRole,
	}

	// Update user limits past midnight
	go func() {
		ctx := context.Background()
		if err := app.updateUserLimits(ctx); err != nil {
			logger.Error("error updating user limit", "error", err.Error())
		}
	}()

	err = app.run(app.mount())
	if err != nil {
		logger.Error("error starting server", "error", err.Error())
		log.Fatal(err)
	}
}
