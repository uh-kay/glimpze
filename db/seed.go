package db

import (
	"context"
	"log"

	"github.com/uh-kay/glimpze/store"
)

func Seed(queries store.Storage) {
	createAdminAccount(queries)
}

func createAdminAccount(queries store.Storage) {
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

	if err := queries.Users.Create(context.Background(), &user); err != nil {
		log.Println(err)
	}
}
