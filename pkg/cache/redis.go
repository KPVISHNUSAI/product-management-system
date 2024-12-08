package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func NewRedisCache(addr string, password string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr, // format should be "host:port"
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *RedisCache) BatchGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for key, cmd := range cmds {
		data, err := cmd.Result()
		if err == nil {
			result[key] = data
		}
	}

	return result, nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
