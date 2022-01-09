package registry

import (
	"sync"

	"github.com/dtm-labs/dtm/dtmsvr/storage"
)

// SingletonFactory is the factory to build store in SINGLETON pattern.
type SingletonFactory struct {
	once sync.Once

	store storage.Store

	creatorFunction func() storage.Store
}

// GetStorage implement the StorageFactory.GetStorage
func (f *SingletonFactory) GetStorage() storage.Store {
	f.once.Do(func() {
		f.store = f.creatorFunction()
	})

	return f.store
}
