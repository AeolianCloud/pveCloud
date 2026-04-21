package cache

import (
	"strings"

	"github.com/go-redis/redis/v8"
)

func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: strings.TrimSpace(addr),
	})
}
