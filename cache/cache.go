package cache

import (
	"hash/fnv"
	"sync"
	"time"
)

var (
	shards = 64
)

type shard struct {
	Entries map[string]interface{}
	Stats   *Stats
	sync.RWMutex
}

type Stats struct {
	Hits    int       `json:"hits"`
	Misses  int       `json:"misses"`
	Set     int       `json:"set"`
	Removed int       `json:"removed"`
	Uptime  time.Time `json:"uptime"`
}

type Cache []*shard

func New() Cache {
	c := make(Cache, shards)
	for i := 0; i < shards; i++ {
		c[i] = NewShard()
	}
	return c
}

func NewShard() *shard {
	return &shard{
		Entries: make(map[string]interface{}),
		Stats:   &Stats{Uptime: time.Now().UTC()},
	}
}

func (c *Cache) Set(key string, value interface{}) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()
	e := value
	s.Entries[key] = e
	s.Stats.Set++
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl int) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()
	e := value
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

func (c *Cache) getShard(key string) *shard {
	h := fnv.New32()
	h.Write([]byte(key))
	return (*c)[uint(h.Sum32())%uint(shards)]
}

func (c *Cache) Get(key string) (interface{}, bool) {
	s := c.getShard(key)
	s.RLock()
	defer s.RUnlock()

	v, ok := s.Entries[key]

	if ok {
		s.Stats.Hits++
	} else {
		s.Stats.Misses++
	}

	return v, ok
}

func (c *Cache) Remove(key string) {
	s := c.getShard(key)
	s.Lock()
	defer s.Unlock()

	delete(s.Entries, key)
	s.Stats.Removed++
}

func (c *Cache) GetStats() *Stats {
	s := Stats{}

	for i, shrd := range *c {
		if i == 0 {
			s.Uptime = shrd.Stats.Uptime
		}
		s.Hits += shrd.Stats.Hits
		s.Misses += shrd.Stats.Misses
		s.Set += shrd.Stats.Set
		s.Removed += shrd.Stats.Removed
	}

	return &s
}
