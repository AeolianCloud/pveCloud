package main

import (
	"log"
	"net/http"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app, err := bootstrap.NewPublicApp(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Server().ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
