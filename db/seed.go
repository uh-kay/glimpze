package db

import (
	"context"
	"log"

	"newsdrop.org/store"
)

func Seed(queries store.Storage) {
	createAdminAccount(queries)
	createUserAccount(queries)
	createTags(queries)
	createPosts(queries)
	createComments(queries)
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
			Title:    "Fedora 43 removed X11 support",
			Content:  "Fedora team said they are removing X11 support in Fedora 43. This means you cannot use X11 anymore without compiling from source code.",
			UserID:   2,
			Username: "npc",
			Tags:     []string{"fedora-linux", "linux"},
		},
	}
	for _, post := range posts {
		newPost, err := queries.Posts.Create(context.Background(), post.Title, post.Content, post.UserID)
		if err != nil {
			log.Println(err)
		}

		for _, tag := range post.Tags {
			_, err := queries.PostTags.Create(context.Background(), newPost.ID, tag)
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func createTags(q store.Storage) {
	tags := []string{"newsdrop", "go", "tech", "fedora-linux", "linux"}

	for _, tag := range tags {
		_, err := q.Tags.Create(context.Background(), tag)
		if err != nil {
			log.Println(err)
		}
	}
}

func createComments(q store.Storage) {
	comments := []store.Comment{
		{
			PostID:  1,
			UserID:  2,
			Content: "Congratz!!",
		},
		{
			PostID:  1,
			UserID:  3,
			Content: "I hope Newsdrop bring a positive impact to the world",
		},
		{
			PostID:  2,
			UserID:  1,
			Content: "I'm excited to see how Green Tea garbage collector going to perform in real world scenario",
		},
	}

	for _, comment := range comments {
		_, err := q.Comments.Create(context.Background(), comment.Content, comment.UserID, comment.PostID)
		if err != nil {
			log.Println(err)
		}
	}
}
