package cache

import (
	"time"

	"github.com/go-redis/redis"

	"github.com/antik9/social-net/internal/config"
)

type Cache struct {
	client *redis.Client
}

var (
	RedisCache = Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     config.Conf.Redis.Address,
			Password: config.Conf.Redis.Password,
			DB:       0,
		}),
	}
)

func (c *Cache) CacheFeedPage(userId, page string) {
	c.client.Set(userId, page, time.Second*60)
}

func (c *Cache) GetFeedPage(userId string) string {
	return c.client.Get(userId).Val()
}
