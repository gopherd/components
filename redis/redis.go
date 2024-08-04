package redis

import (
	"github.com/go-redis/redis/v8"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/redis"

// Options represents the options of the component.
type Options struct {
	Source string `json:"source"`
}

// Component defines the redis component API
type Component interface {
	// Client returns the redis client
	Client() *redis.Client
	// Key returns the key with prefix (aka namespace)
	Key(key string) string
}
