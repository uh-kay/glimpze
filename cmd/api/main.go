package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/uh-kay/xed/env"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":3000"),
		env:  env.GetString("ENV", "dev"),
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app := &application{
		logger: logger,
		config: cfg,
	}

	err := app.run(app.mount())
	if err != nil {
		logger.Error("error starting server", "error", err.Error())
		log.Fatal(err)
	}
}
