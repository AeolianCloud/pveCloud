package main

import (
	"log"

	"pvecloud/backend/internal/server"
)

// 小薄层，保留 cmd/server 目录作为可执行入口。
func main() {
	if err := server.Run("config/config.yaml"); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
