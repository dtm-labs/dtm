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

// StorageFactory is factory to get storage instance.
type StorageFactory interface {
	// GetStorage will return the Storage instance.
	GetStorage() storage.Store
}

var storeFactorys = map[string]StorageFactory{
	"boltdb": &SingletonFactory{
		creatorFunction: func() storage.Store {
			return boltdb.NewStore(conf.Store.DataExpire, conf.RetryInterval)
		},
	},
	"redis": &SingletonFactory{
		creatorFunction: func() storage.Store {
			return &redis.Store{}
		},
	},
	"mysql": &SingletonFactory{
		creatorFunction: func() storage.Store {
			return &sql.Store{}
		},
	},
	"postgres": &SingletonFactory{
		creatorFunction: func() storage.Store {
			return &sql.Store{}
		},
	},
}

// GetStore returns storage.Store
func GetStore() storage.Store {
	return storeFactorys[conf.Store.Driver].GetStorage()
}

// WaitStoreUp wait for db to go up
func WaitStoreUp() {
	for err := GetStore().Ping(); err != nil; err = GetStore().Ping() {
		time.Sleep(3 * time.Second)
	}
}
