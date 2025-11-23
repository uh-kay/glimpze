package db

import (
	"context"
	"log"

	"newsdrop.org/store"
)

func Seed(queries store.Storage) {
	createAdminAccount(queries)
	createUserAccount(queries)
	createPosts(queries)
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

	err = queries.WithTx(context.Background(), func(s *store.Storage) error {
		if err := queries.Users.Create(context.Background(), &user); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func createUserAccount(queries store.Storage) {
	userRole, err := queries.Roles.GetByName(context.Background(), "user")
	if err != nil {
		log.Println(err)
	}

	users := []store.User{
		{
			Name:        "npc",
			DisplayName: "npc",
			Email:       "npc@example.com",
			Role:        *userRole,
		},
		{
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

		err = queries.WithTx(context.Background(), func(s *store.Storage) error {
			if err := queries.Users.Create(context.Background(), &user); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func createPosts(queries store.Storage) {
	posts := []store.Post{
		{
			Title:    "Hello world!",
			Content:  "Drop finally launched after months in development. We would like to thank everyone who help with Drop development.",
			UserID:   1,
			Username: "admin",
			Tags:     []string{"newsdrop"},
		},
		{
			Title:    "Go 1.25 released",
			Content:  "Last week the Go team announced Go 1.25, a major release after 6 months in development. This latest Go version bring a new experimental Green Tea garbage collector.",
			UserID:   2,
			Username: "npc",
			Tags:     []string{"go", "tech"},
		},
		{
			Title:    "Fedora 43 remove X11 support",
			Content:  "Fedora team said they are removing X11 support in Fedora 43. This means you cannot use X11 anymore without compiling from source code.",
			UserID:   2,
			Username: "npc",
			Tags:     []string{"fedora-linux", "linux"},
		},
	}
	for i := range posts {
		_, err := queries.Posts.Create(context.Background(), posts[i].Title, posts[i].Content, posts[i].UserID)
		if err != nil {
			log.Println(err)
		}
	}

}
