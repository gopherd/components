package component

import (
	"context"
	"strings"

	goredis "github.com/go-redis/redis/v8"
	"github.com/gopherd/doge/erron"
	"github.com/gopherd/mosaic/component"
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

func (com *redisComponent) Init(ctx context.Context, entity component.Entity) error {
	if client, opt, err := redisapi.NewClient(com.Options().Source); err != nil {
		return erron.Throw(err)
	} else {
		com.client = client
		com.prefix = opt.Prefix
		if !strings.HasSuffix(com.prefix, ".") {
			com.prefix += "."
		}
	}
	return nil
}
