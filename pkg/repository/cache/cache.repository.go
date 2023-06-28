package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/isd-sgcu/rpkm66-file/internal/repository/cache"
)

type CacheRepository interface {
	SaveCache(string, interface{}, int) error
	GetCache(string, interface{}) error
}

func NewRepository(client *redis.Client) CacheRepository {
	return cache.NewRepository(client)
}
