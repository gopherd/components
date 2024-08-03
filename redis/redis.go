package redis

import (
	"github.com/go-redis/redis/v8"
)

const ComponentName = "redis"

type Options struct {
	Source string `json:"source"`
}

// Component defines the redis component API
type Component interface {
	// Client returns the redis client
	Client() *redis.Client
	// Key returns the key with prefix (or namespace)
	Key(key string) string
}
