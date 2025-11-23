package main

import (
	"log"

	"newsdrop.org/db"
	"newsdrop.org/env"
	"newsdrop.org/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://root:password@localhost:5432/newsdrop?sslmode=disable")
	conn, err := db.New(
		addr,
		30,
		"15m",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store)
}
