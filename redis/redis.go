package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gopherd/core/component"

	"github.com/gopherd/components/redis/redisapi"
)

// Name is the unique identifier for the redis component.
const Name = "github.com/gopherd/components/redis"

func init() {
	component.Register(Name, func() component.Component {
		return &redisComponent{}
	})
}

// Ensure redisComponent implements redisapi.Component interface.
var _ redisapi.Component = (*redisComponent)(nil)

// Options defines the configuration options for the redis component.
type Options struct {
	// The network type, either tcp or unix.
	// Default is tcp.
	Network string
	// host:port address.
	Addr string

	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string

	// Database to be selected after connecting to the server.
	DB int

	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration

	// Type of connection pool.
	// true for FIFO pool, false for LIFO pool.
	// Note that fifo has higher overhead compared to lifo.
	PoolFIFO bool
	// Maximum number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration
}

type redisComponent struct {
	component.BaseComponent[Options]

	client *redis.Client
}

func (c *redisComponent) Init(ctx context.Context) error {
	options := c.Options()
	client := redis.NewClient(&redis.Options{
		Network:            options.Network,
		Addr:               options.Addr,
		Username:           options.Username,
		Password:           options.Password,
		DB:                 options.DB,
		MaxRetries:         options.MaxRetries,
		MinRetryBackoff:    options.MinRetryBackoff,
		MaxRetryBackoff:    options.MaxRetryBackoff,
		DialTimeout:        options.DialTimeout,
		ReadTimeout:        options.ReadTimeout,
		WriteTimeout:       options.WriteTimeout,
		PoolFIFO:           options.PoolFIFO,
		PoolSize:           options.PoolSize,
		MinIdleConns:       options.MinIdleConns,
		MaxConnAge:         options.MaxConnAge,
		PoolTimeout:        options.PoolTimeout,
		IdleTimeout:        options.IdleTimeout,
		IdleCheckFrequency: options.IdleCheckFrequency,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *redisComponent) Uninit(ctx context.Context) error {
	return c.client.Close()
}

func (c *redisComponent) Client() *redis.Client {
	return c.client
}
