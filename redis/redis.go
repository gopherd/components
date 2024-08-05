package redis

import (
	"context"
	"strings"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gopherd/core/component"
	redisapi "github.com/gopherd/redis/api"

	api "github.com/gopherd/components/redis/api"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/redis"

// Options represents the options of the component.
type Options struct {
	Source string `json:"source"`
}

var _ api.Component = (*redisComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &redisComponent{}
	})
}

type redisComponent struct {
	component.BaseComponent[Options]

	client *goredis.Client
	prefix string
}

func (com *redisComponent) Init(ctx context.Context) error {
	if client, opt, err := redisapi.NewClient(com.Options().Source); err != nil {
		return err
	} else {
		if err := client.Ping(ctx).Err(); err != nil {
			return err
		}
		com.client = client
		com.prefix = opt.Prefix
		if !strings.HasSuffix(com.prefix, ".") {
			com.prefix += "."
		}
	}
	return nil
}

func (com *redisComponent) Uninit(ctx context.Context) error {
	return com.client.Close()
}

func (com *redisComponent) Client() *goredis.Client {
	return com.client
}

func (com *redisComponent) Key(key string) string {
	return com.prefix + key
}
