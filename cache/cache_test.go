package cache

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestSetEntry(t *testing.T) {
	key := "testKey"
	value := "testValue"

	c := New()

	c.Set(key, value)

	v, ok := c.Get(key)

	if !ok {
		t.Error("Could not find test element in cache.")
		t.Fail()
	}

	str, ok := v.(string)
	if !ok {
		t.Error("Found value could not be asserted as string.")
		t.Fail()
	}

	if str != value {
		t.Error("Expected", value, "got", str)
		t.Fail()
	}
}

type testType struct {
	Val1 string
	Val2 int
}

func TestSetTyped(t *testing.T) {
	key := "testKey"
	value := testType{
		Val1: "testValue",
		Val2: 42,
	}

	c := New()

	c.Set(key, value)

	v, ok := c.Get(key)

	if !ok {
		t.Error("Could not find test element in cache.")
		t.Fail()
	}

	tt, ok := v.(testType)
	if !ok {
		t.Error("Found value could not be asserted as testType.")
		t.Fail()
	}

	if tt.Val1 != value.Val1 {
		t.Error("Expected", value.Val1, "got", tt.Val1)
		t.Fail()
	}

	if tt.Val2 != value.Val2 {
		t.Error("Expected", value.Val2, "got", tt.Val2)
		t.Fail()
	}
}

func TestSetWithTTL(t *testing.T) {
	key := "testKey"
	value := "testValue"
	ttl := 2

	c := New()

	c.SetWithTTL(key, value, ttl)

	v, ok := c.Get(key)

	if !ok {
		t.Error("Could not find test element in cache.")
		t.Fail()
	}

	str, ok := v.(string)
	if !ok {
		t.Error("Found value could not be asserted as string.")
		t.Fail()
	}

	if str != value {
		t.Error("Expected", value, "got", str)
		t.Fail()
	}

	time.Sleep(3 * time.Second)

	v, ok = c.Get(key)

	if ok {
		t.Error("Element should have been removed.")
		t.Fail()
	}
}

func TestStats(t *testing.T) {
	key := "testKey"
	value := "testValue"

	c := New()

	for i := 0; i < 100; i++ {
		c.Set(key+strconv.Itoa(i), value)
	}

	s := c.GetStats()

	if s == nil {
		t.Error("Stats should not be nil")
		t.Fail()
	}

	if s.Set != 100 {
		t.Errorf("Expected 100 values to be set. Got %d", s.Set)
		t.Fail()
	}

	for i := 0; i < 100; i++ {
		c.Get(key + strconv.Itoa(i))
	}

	s = c.GetStats()

	if s == nil {
		t.Error("Stats should not be nil")
		t.Fail()
	}

	if s.Hits != 100 {
		t.Errorf("Expected 100 cache hits. Got %d", s.Hits)
		t.Fail()
	}

	for i := 0; i < 100; i++ {
		c.Remove(key + strconv.Itoa(i))
	}

	s = c.GetStats()

	if s == nil {
		t.Error("Stats should not be nil")
		t.Fail()
	}

	if s.Removed != 100 || s.Removed != s.Set {
		t.Errorf("Expected 100 values to be removed. Got %d", s.Removed)
		t.Fail()
	}

	for i := 0; i < 100; i++ {
		c.Get(key + strconv.Itoa(i))
	}

	s = c.GetStats()

	if s == nil {
		t.Error("Stats should not be nil")
		t.Fail()
	}

	if s.Misses != 100 {
		t.Errorf("Expected 100 cache misses. Got %d", s.Misses)
		t.Fail()
	}

}

func BenchmarkCache(b *testing.B) {
	c := New()

	for i := 0; i < b.N; i++ {
		value := testType{
			Val1: "testValue",
			Val2: 42,
		}
		c.Set(strconv.Itoa(i), value)
		v, ok := c.Get(strconv.Itoa(i))
		if ok {
			_, ok := v.(testType)
			if ok {
				// done
			}
		}
	}
}

func BenchmarkCacheParallel(b *testing.B) {
	c := New()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			value := testType{
				Val1: "testValue",
				Val2: 42,
			}
			c.Set(strconv.Itoa(rand.Intn(int(^uint(0)>>1))), value)
			v, ok := c.Get(strconv.Itoa(rand.Intn(int(^uint(0) >> 1))))
			if ok {
				_, ok := v.(testType)
				if ok {
					// done
				}
			}
		}
	})
}

func BenchmarkCacheTTL(b *testing.B) {
	c := New()

	for i := 0; i < b.N; i++ {
		value := testType{
			Val1: "testValue",
			Val2: 42,
		}
		c.SetWithTTL(strconv.Itoa(i), value, 1)
		v, ok := c.Get(strconv.Itoa(i))
		if ok {
			_, ok := v.(testType)
			if ok {
				// done
			}
		}
	}
}

func BenchmarkCacheTTLParallel(b *testing.B) {
	c := New()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			value := testType{
				Val1: "testValue",
				Val2: 42,
			}
			c.SetWithTTL(strconv.Itoa(rand.Intn(int(^uint(0)>>1))), value, 1)
			v, ok := c.Get(strconv.Itoa(rand.Intn(int(^uint(0) >> 1))))
			if ok {
				_, ok := v.(testType)
				if ok {
					// done
				}
			}
		}
	})
}
