package storage

import (
	"github.com/go-redis/redis/v8"
	"github.com/yedf/dtm/common"
)

var config = &common.Config

func dbGet() *common.DB {
	return common.DbGet(config.Store.GetDBConf())
}

func redisGet() *redis.Client {
	return common.RedisGet()
}
