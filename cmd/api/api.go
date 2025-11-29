package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"newsdrop.org/env"
	"newsdrop.org/mailer"
	"newsdrop.org/storage"
	"newsdrop.org/store"
	"newsdrop.org/store/cache"
)

type application struct {
	config      config
	store       store.Storage
	logger      *slog.Logger
	db          *pgxpool.Pool
	cache       cache.Storage
	storage     *storage.R2Client
	defaultRole *store.Role
	mailer      mailer.Client
	wg          sync.WaitGroup
}

type config struct {
	addr         string
	env          string
	dbConfig     dbConfig
	valkeyCfg    valkeyCfg
	r2Cfg        r2Cfg
	rateLimitCfg rateLimitCfg
	mailCfg      mailCfg
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

type rateLimitCfg struct {
	requestCount int
	windowLength time.Duration
}

type mailCfg struct {
	apiKey    string
	fromEmail string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("FRONTEND_URL", "http://localhost:5173")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(httprate.Limit(
		app.config.rateLimitCfg.requestCount,
		app.config.rateLimitCfg.windowLength,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			app.rateLimitExceededResponse(w, r, "rate limit exceeded, retry after: 1 minute")
		})),
	)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthcheck)

		r.Group(func(r chi.Router) {
			r.Use(app.optionalAuthMiddleware)
			r.Get("/", app.userFeed)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", app.login)
			r.Post("/register", app.register)
			r.Patch("/activate/", app.activateUser)
			r.Post("/token/refresh", app.refreshToken)
			r.Post("/logout", app.logout)
		})

		r.Route("/posts", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(app.AuthMiddleware)
				r.Post("/", app.createPost)
				r.Get("/users/{userID}", app.getPostByUserID)
				r.Get("/users/", app.getPostByUserID)
			})

			r.Post("/upload", app.uploadPostFiles)

			r.Route("/{postID}", func(r chi.Router) {
				r.Get("/", app.getPost)

				r.Group(func(r chi.Router) {
					// r.Use(app.AuthMiddleware)
					r.Use(app.postContextMiddleware)

					r.Patch("/", app.checkPostOwnership("moderator", app.updatePost))
					r.Delete("/", app.checkPostOwnership("admin", app.deletePost))

					r.Route("/likes", func(r chi.Router) {
						r.Post("/", app.addLike)
						r.Delete("/", app.removeLike)
					})

					r.Route("/tags", func(r chi.Router) {
						r.Post("/", app.addTag)
						r.Get("/", app.listTag)
						r.Delete("/{tagID}", app.removeTag)
					})

					r.Route("/comments", func(r chi.Router) {
						r.Get("/", app.listComment)
						r.Post("/", app.createComment)
						r.Get("/{commentID}", app.getComment)
						r.Patch("/{commentID}", app.updateComment)
						r.Delete("/{commentID}", app.deleteComment)
					})
				})
			})
		})

		r.Route("/tags", func(r chi.Router) {
			r.Get("/{tagName}/posts", app.getPostByTag)

			r.Group(func(r chi.Router) {
				r.Use(app.AuthMiddleware)
				r.Post("/", app.checkResourceAccess("moderator", app.createTag))
				r.Get("/{tagID}", app.getTag)
				r.Delete("/{tagID}", app.checkResourceAccess("moderator", app.deleteTag))
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			// r.Patch("/{userName}", app.checkResourceAccess("admin", app.updateUserRole))
			r.Get("/{userID}", app.profile)
			r.Get("/", app.profile)
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
