package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uh-kay/glimpze/store"
)

func Seed(db *pgxpool.Pool, queries store.Storage) {
	createAdminAccount(db, queries)
	createUserAccount(db, queries)
}

func createAdminAccount(db *pgxpool.Pool, queries store.Storage) {
	adminRole, err := queries.Roles.GetByName(context.Background(), "admin")
	if err != nil {
		log.Println(err)
	}

	user := store.User{
		Name:        "admin",
		DisplayName: "admin",
		Email:       "admin@example.com",
		Role:        *adminRole,
	}

	if err := user.Password.Set("password"); err != nil {
		log.Println(err)
	}

	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Println(err)
	}
	defer tx.Rollback(context.Background())

	if err := queries.Users.Create(context.Background(), tx, &user); err != nil {
		log.Println(err)
	}

	if _, err := queries.UserLimits.Create(context.Background(), tx, user.ID); err != nil {
		log.Println(err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		log.Println(err)
	}
}

func createUserAccount(db *pgxpool.Pool, queries store.Storage) {
	userRole, err := queries.Roles.GetByName(context.Background(), "user")
	if err != nil {
		log.Println(err)
	}

	users := []store.User{
		store.User{
			Name:        "npc",
			DisplayName: "npc",
			Email:       "npc@example.com",
			Role:        *userRole,
		},
		store.User{
			Name:        "normie",
			DisplayName: "normie",
			Email:       "normie@example.com",
			Role:        *userRole,
		},
	}

	for _, user := range users {
		if err := user.Password.Set("password"); err != nil {
			log.Println(err)
		}

		tx, err := db.Begin(context.Background())
		if err != nil {
			log.Println(err)
		}
		defer tx.Rollback(context.Background())

		if err := queries.Users.Create(context.Background(), tx, &user); err != nil {
			log.Println(err)
		}

		if _, err := queries.UserLimits.Create(context.Background(), tx, user.ID); err != nil {
			log.Println(err)
		}

		if err := tx.Commit(context.Background()); err != nil {
			log.Println(err)
		}
	}
}
