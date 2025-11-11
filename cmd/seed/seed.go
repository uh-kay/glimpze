package main

import (
	"log"

	"github.com/uh-kay/glimpze/db"
	"github.com/uh-kay/glimpze/env"
	"github.com/uh-kay/glimpze/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://root:password@localhost:5432/glimpze?sslmode=disable")
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
