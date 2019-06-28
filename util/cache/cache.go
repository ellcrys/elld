package cache

import (
	"time"

	"github.com/ellcrys/elld/util"

	"github.com/hashicorp/golang-lru"
)

// CacheItemRemovalInterval is the time
// interval when expired cache item are
// checked and remove.
var CacheItemRemovalInterval = 5 * time.Second

// Sec returns current time + sec
func Sec(sec int) time.Time {
	return time.Now().Add(time.Duration(sec) * time.Second)
}

type cacheValue struct {
	value      interface{}
	expiration time.Time
}

// Cache represents a container for storing objects in memory.
// Internally it uses an LRU cache allowing older items to be
// removed eventually. It also supports expiring items that
// are by default, lazily removed during new additions. To
// support active expiry checks and removal, use NewActiveCache.
type Cache struct {
	container *lru.Cache
}

// NewCache creates a new cache
func NewCache(capacity int) *Cache {
	cache := new(Cache)
	cache.container, _ = lru.New(capacity)
	return cache
}

// NewActiveCache creates a new cache instance
// and begins active item expiration checks
// and removal.
func NewActiveCache(capacity int) *Cache {
	cache := NewCache(capacity)
	go func() {
		for {
			<-time.NewTicker(CacheItemRemovalInterval).C
			cache.removeExpired()
		}
	}()
	return cache
}

// Add adds an item to the cache. It the cache becomes
// full, the oldest item is deleted to make room for the new value.
func (c *Cache) Add(key, val interface{}) {
	c.add(key, val, time.Time{})
}

// AddMulti adds multiple values into
// the cache. The values are serialized
// and used as the key.
func (c *Cache) AddMulti(exp time.Time, values ...interface{}) {
	c.removeExpired()
	valueHex := util.ToHex(util.ObjectToBytes(values))
	c.AddWithExp(valueHex, values, exp)
}

// HasMulti checks whether a multiple
// valued serialized key exist in the cache..
func (c *Cache) HasMulti(values ...interface{}) bool {
	c.removeExpired()
	valueHex := util.ToHex(util.ObjectToBytes(values))
	return c.Has(valueHex)
}

func (c *Cache) add(key, val interface{}, expTime time.Time) {
	c.removeExpired()
	c.container.Add(key, &cacheValue{
		value:      val,
		expiration: expTime,
	})
}

// AddWithExp adds an item to the cache with an explicit expiration time.
// Expired items are removed lazily - Whenever Add or AddWithExp are called.
// An item with an expiry time does not need to be the oldest in the cache before it is removed.
func (c *Cache) AddWithExp(key, val interface{}, expTime time.Time) {
	c.add(key, val, expTime)
}

// Peek gets an item without updating the newness of the item
func (c *Cache) Peek(key interface{}) interface{} {
	v, _ := c.container.Peek(key)
	if v == nil {
		return nil
	}
	return v.(*cacheValue).value
}

// Get gets an item and updates the newness of the item
func (c *Cache) Get(key interface{}) interface{} {
	v, _ := c.container.Get(key)
	if v == nil {
		return nil
	}
	return v.(*cacheValue).value
}

// removeExpired removes expired items
func (c *Cache) removeExpired() {
	for _, k := range c.container.Keys() {

		cVal, ok := c.container.Peek(k)
		if !ok {
			continue
		}

		if cVal.(*cacheValue).expiration.IsZero() {
			continue
		}

		if time.Now().After(cVal.(*cacheValue).expiration) {
			c.container.Remove(k)
		}
	}
}

// Keys returns all keys in the cache
func (c *Cache) Keys() []interface{} {
	return c.container.Keys()
}

// Remove removes an item from the cache
func (c *Cache) Remove(key interface{}) {
	c.container.Remove(key)
}

// Has checks whether an item is in the cache without updating the newness of the item
func (c *Cache) Has(key interface{}) bool {
	return c.container.Contains(key)
}

// Len returns the length of the cache
func (c *Cache) Len() int {
	return c.container.Len()
}
