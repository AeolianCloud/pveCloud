package main

import (
	"flag"
	"log"

	"pvecloud/backend/internal/server"
)

// 在仓库根提供直接启动入口，支持 `go run .`。
func main() {
	cfg := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	if err := server.Run(*cfg); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
