package cache

import (
	"hash/fnv"
	"sync"
	"time"
)

var (
	shards = 64
)

type entry struct {
	value interface{}
	exit  chan struct{}
}

type shard struct {
	Entries map[string]entry
	Stats   *Stats
	sync.RWMutex
}

// Stats represents access statistics of a Cache.
type Stats struct {
	Hits    int       `json:"hits"`
	Misses  int       `json:"misses"`
	Set     int       `json:"set"`
	Removed int       `json:"removed"`
	Uptime  time.Time `json:"uptime"`
}

// Cache is a thread safe structure to store and retrieve arbitrary values.
type Cache []*shard

// New returns a reference to a new Cache
func New() *Cache {
	c := make(Cache, shards)
	for i := 0; i < shards; i++ {
		c[i] = newShard()
	}
	return &c
}

func newShard() *shard {
	return &shard{
		Entries: make(map[string]entry),
		Stats:   &Stats{Uptime: time.Now().UTC()},
	}
}

// Set stores the value with the given key.
func (c *Cache) Set(key string, value interface{}) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()

	e, ok := s.Entries[key]
	// make sure to exit the ttl go routine of a previously stored value
	// before overwriting it
	if ok && e.exit != nil {
		e.exit <- struct{}{}
	}

	e.value = value
	s.Entries[key] = e

	s.Stats.Set++
}

// SetWithTTL stores the value with the given key and removes it automatically after
// ttl seconds.
func (c *Cache) SetWithTTL(key string, value interface{}, ttl int) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()

	e, ok := s.Entries[key]
	// make sure to exit the ttl go routine of a previously stored value
	// before overwriting it
	if ok && e.exit != nil {
		e.exit <- struct{}{}
	}

	e.value = value
	e.exit = make(chan struct{}, 1)

	s.Entries[key] = e
	s.Stats.Set++

	timeout := time.Tick(time.Second * time.Duration(ttl))
	// wait for the timeout concurrently
	go func() {
		for {
			select {
			case <-timeout:
				c.Remove(key)
				return
			case <-e.exit:
				c.Remove(key)
				return
			}
		}
	}()
}

func (c *Cache) getShard(key string) *shard {
	h := fnv.New32()
	h.Write([]byte(key))
	return (*c)[uint(h.Sum32())%uint(shards)]
}

// Get retrieves a value stored with a specific key. If no value is available
// nil and false will be returned.
func (c *Cache) Get(key string) (interface{}, bool) {
	s := c.getShard(key)
	s.RLock()
	defer s.RUnlock()

	v, ok := s.Entries[key]

	if ok {
		s.Stats.Hits++
		return v.value, true
	}

	s.Stats.Misses++

	return nil, ok
}

// Remove deletes a value stored with the given key from the cache.
// In case no value exists no action is performed.
func (c *Cache) Remove(key string) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()

	e, ok := s.Entries[key]
	if ok {
		if e.exit != nil {
			e.exit <- struct{}{}
		}
		delete(s.Entries, key)
		s.Stats.Removed++
	}
}

func (c *Cache) len() int {
	return len(*c)
}

func (c *Cache) shard(n int) *shard {
	return (*c)[n]
}

// GetStats returns Stats for this cache instance.
func (c *Cache) GetStats() *Stats {
	s := Stats{}

	for i := 0; i < c.len(); i++ {
		shrd := c.shard(i)
		shrd.Lock()

		if i == 0 {
			s.Uptime = shrd.Stats.Uptime
		}
		s.Hits += shrd.Stats.Hits
		s.Misses += shrd.Stats.Misses
		s.Set += shrd.Stats.Set
		s.Removed += shrd.Stats.Removed

		shrd.Unlock()
	}

	return &s
}
