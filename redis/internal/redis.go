package internal

import (
	"context"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gopherd/core/component"

	"github.com/gopherd/components/redis"
)

func init() {
	component.Register(redis.Name, func() component.Component {
		return &RedisComponent{}
	})
}

type RedisComponent struct {
	component.BaseComponent[redis.Options]

	client *goredis.Client
}

func (c *RedisComponent) Init(ctx context.Context) error {
	options := c.Options()
	client := goredis.NewClient(&goredis.Options{
		Network:            options.Network,
		Addr:               options.Addr,
		Username:           options.Username,
		Password:           options.Password,
		DB:                 options.DB,
		MaxRetries:         options.MaxRetries,
		MinRetryBackoff:    options.MinRetryBackoff.Value(),
		MaxRetryBackoff:    options.MaxRetryBackoff.Value(),
		DialTimeout:        options.DialTimeout.Value(),
		ReadTimeout:        options.ReadTimeout.Value(),
		WriteTimeout:       options.WriteTimeout.Value(),
		PoolFIFO:           options.PoolFIFO,
		PoolSize:           options.PoolSize,
		MinIdleConns:       options.MinIdleConns,
		MaxConnAge:         options.MaxConnAge.Value(),
		PoolTimeout:        options.PoolTimeout.Value(),
		IdleTimeout:        options.IdleTimeout.Value(),
		IdleCheckFrequency: options.IdleCheckFrequency.Value(),
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *RedisComponent) Uninit(ctx context.Context) error {
	return c.client.Close()
}

// Ensure redisComponent implements redisapi.Component interface.
var _ redis.Component = (*RedisComponent)(nil)

// Client implements redis.Component Client method.
func (c *RedisComponent) Client() *goredis.Client {
	return c.client
}
