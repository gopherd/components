package component

import (
	"context"
	"strings"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gopherd/core/component"
	redisapi "github.com/gopherd/redis/api"

	"github.com/gopherd/components/redis"
)

var _ redis.Component = (*redisComponent)(nil)

func init() {
	component.Register(redis.ComponentName, func() component.Component {
		return &redisComponent{}
	})
}

type redisComponent struct {
	component.BaseComponent[redis.Options]

	client *goredis.Client
	prefix string
}

func (com *redisComponent) Init(ctx context.Context) error {
	if client, opt, err := redisapi.NewClient(com.Options().Source); err != nil {
		return err
	} else {
		com.client = client
		com.prefix = opt.Prefix
		if !strings.HasSuffix(com.prefix, ".") {
			com.prefix += "."
		}
	}
	return nil
}

func (com *redisComponent) Client() *goredis.Client {
	return com.client
}

func (com *redisComponent) Key(key string) string {
	return com.prefix + key
}
