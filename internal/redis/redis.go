package redis

import (
	"context"
	"github.com/TiktokCommence/authService/internal/conf"
	"github.com/go-redis/redis"
	"github.com/google/wire"
	"time"
)

// ProviderSet is redis providers.
var ProviderSet = wire.NewSet(NewRedisDB, NewCache)

// NewRedisDB 连接redis
func NewRedisDB(c *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.GetAddr(),
		ReadTimeout:  c.Redis.GetReadTimeout().AsDuration(),
		WriteTimeout: c.Redis.GetWriteTimeout().AsDuration(),
		DB:           0,
		Password:     c.Redis.GetPassword(),
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		panic(err)
	}
	return rdb
}

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}
func (c *Cache) SetKV(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	return c.client.Set(key, value, expire).Err()
}
func (c *Cache) GetValue(ctx context.Context, key string) (interface{}, error) {
	return c.client.Get(key).Result()
}
func (c *Cache) ExistKey(ctx context.Context, key string) bool {
	exists, err := c.client.Exists(key).Result()
	if err != nil {
		return false
	}
	if exists > 0 {
		return true
	}
	return false
}

func (c *Cache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(key).Result()
}
