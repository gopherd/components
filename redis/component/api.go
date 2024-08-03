package component

import goredis "github.com/go-redis/redis/v8"

func (com *redisComponent) Client() *goredis.Client {
	return com.client
}

func (com *redisComponent) Key(key string) string {
	return com.prefix + key
}
