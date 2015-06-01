package cache

import (
	"hash/fnv"
	"sync"
	"time"
)

var shards = 64

type CacheEntry struct {
	Value interface{}
}

type cacheShard struct {
	Entries map[string]CacheEntry
	sync.RWMutex
}

type Cache []*cacheShard

func New() Cache {
	c := make(Cache, shards)
	for i := 0; i < shards; i++ {
		c[i] = &cacheShard{Entries: make(map[string]CacheEntry)}
	}
	return c
}

func (c *Cache) Set(key string, value interface{}) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()
	e := CacheEntry{Value: value}
	s.Entries[key] = e
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl int) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()
	e := CacheEntry{
		Value: value,
	}

	timeout := time.Tick(time.Second * time.Duration(ttl))
	s.Entries[key] = e
	go func() {
		for {
			select {
			case <-timeout:
				c.Remove(key)
				return
			}
		}
	}()
}

func (c *Cache) getShard(key string) *cacheShard {
	h := fnv.New32()
	h.Write([]byte(key))
	return (*c)[uint(h.Sum32())%uint(shards)]
}

func (c *Cache) Get(key string) (interface{}, bool) {
	s := c.getShard(key)
	s.RLock()
	defer s.RUnlock()

	v, ok := s.Entries[key]
	return v.Value, ok
}

func (c *Cache) Remove(key string) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()

	delete(s.Entries, key)
}
