package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uh-kay/glimpze/storage"
	"github.com/uh-kay/glimpze/store"
	"github.com/uh-kay/glimpze/store/cache"
)

type application struct {
	config  config
	store   store.Storage
	logger  *slog.Logger
	db      *pgxpool.Pool
	cache   cache.Storage
	storage *storage.R2Client
}

type config struct {
	addr      string
	env       string
	dbConfig  dbConfig
	valkeyCfg valkeyCfg
	r2Cfg     r2Cfg
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleTime  string
}

type valkeyCfg struct {
	enabled bool
	addr    string
	pw      string
	db      int
}

type r2Cfg struct {
	bucketName      string
	accountID       string
	accessKeyID     string
	accessKeySecret string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthcheck)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", app.login)
			r.Post("/register", app.register)
			r.Post("/token/refresh", app.refreshToken)

			r.Route("/profile", func(r chi.Router) {
				r.Use(app.AuthMiddleware)
				r.Get("/", app.profile)
			})
		})

		r.Route("/posts", func(r chi.Router) {
			r.Get("/{postID}", app.getPost)

			r.Group(func(r chi.Router) {
				r.Use(app.AuthMiddleware)

				r.Post("/", app.createPost)
				r.Patch("/{postID}", app.updatePost)
				r.Delete("/{postID}", app.deletePost)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Info("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Info("server has started", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Info("server has stopped", "addr", app.config.addr, "env", app.config.env)

	return nil
}
