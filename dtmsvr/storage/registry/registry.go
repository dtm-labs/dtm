package registry

import (
	"time"

	"github.com/dtm-labs/dtm/dtmsvr/config"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
	"github.com/dtm-labs/dtm/dtmsvr/storage/boltdb"
	"github.com/dtm-labs/dtm/dtmsvr/storage/redis"
	"github.com/dtm-labs/dtm/dtmsvr/storage/sql"
)

var conf = &config.Config

var stores map[string]storage.Store = map[string]storage.Store{
	"redis":    &redis.RedisStore{},
	"mysql":    &sql.SqlStore{},
	"postgres": &sql.SqlStore{},
	"boltdb":   &boltdb.BoltdbStore{},
}

func GetStore() storage.Store {
	return stores[conf.Store.Driver]
}

// WaitStoreUp wait for db to go up
func WaitStoreUp() {
	for err := GetStore().Ping(); err != nil; err = GetStore().Ping() {
		time.Sleep(3 * time.Second)
	}
}
