package cache

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coocood/freecache"
)

type CacheStats struct {
	Rooms     uint64
	Nodes     uint64
	Addresses uint64
	HitCount  int64
}

type Cache struct {
	store  *freecache.Cache
	Ticker *time.Timer
}

func NewCache(size int) *Cache {
	location, _ := time.LoadLocation("UTC") // Ensure time is calculated in UTC
	now := time.Now().In(location)
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, location)
	durationUntilNext := next.Sub(now)
	timer := time.NewTimer(durationUntilNext)

	c := &Cache{
		store:  freecache.NewCache(size),
		Ticker: timer,
	}

	go func() {
		for {
			select {
			case t := <-timer.C:
				fmt.Println("Reset Index Stats", t)
				c.store.ResetStatistics()

				// This is good enough...
				timer.Reset(24 * time.Hour)
			}
		}
	}()

	return c
}

func (c *Cache) Set(key []byte, value []byte, ttl int) error {
	return c.store.Set(key, value, ttl)
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	return c.store.Get(key)
}

func (c *Cache) GetStats() (stats CacheStats) {
	iterator := c.store.NewIterator()

	for {
		next := iterator.Next()
		if next == nil {
			break
		}

		key := string(next.Key)
		if strings.HasPrefix(key, "room:") {
			stats.Rooms++
		}
		if strings.HasPrefix(key, "node:") {
			stats.Nodes++

			// get the addresses for this node
			var values []string
			if err := json.Unmarshal(next.Value, &values); err == nil {
				stats.Addresses += uint64(len(values))
			}
		}
	}

	stats.HitCount = c.store.HitCount()

	return
}
