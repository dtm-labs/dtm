package registry

import (
    "time"
    
    "github.com/dtm-labs/dtm/common"
    "github.com/dtm-labs/dtm/dtmsvr/storage"
    "github.com/dtm-labs/dtm/dtmsvr/storage/boltdb"
    "github.com/dtm-labs/dtm/dtmsvr/storage/redis"
    "github.com/dtm-labs/dtm/dtmsvr/storage/sql"
)

var config = &common.Config

var stores map[string]storage.Store = map[string]storage.Store{
    "redis":    &redis.RedisStore{},
    "mysql":    &sql.SqlStore{},
    "postgres": &sql.SqlStore{},
    "boltdb":   &boltdb.BoltdbStore{},
}

// GetStore get storage driver
func GetStore() storage.Store {
    return stores[config.Store.Driver]
}

// WaitStoreUp wait for db to go up
func WaitStoreUp() {
    for err := GetStore().Ping(); err != nil; err = GetStore().Ping() {
        time.Sleep(3 * time.Second)
    }
}
