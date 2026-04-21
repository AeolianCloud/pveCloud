package main

import (
	"log"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
)

func main() {
	if err := bootstrap.NewServer(":8082").ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
