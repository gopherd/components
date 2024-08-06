package redisapi

import "github.com/go-redis/redis/v8"

// Component defines the redis component API
type Component interface {
	// Client returns the redis client
	Client() *redis.Client
}
