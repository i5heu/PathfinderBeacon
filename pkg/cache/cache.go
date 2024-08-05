package cache

import (
	"github.com/coocood/freecache"
)

type Cache struct {
	store *freecache.Cache
}

func NewCache(size int) *Cache {
	return &Cache{
		store: freecache.NewCache(size),
	}
}

func (c *Cache) Set(key []byte, value []byte, ttl int) error {
	return c.store.Set(key, value, ttl)
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	return c.store.Get(key)
}
