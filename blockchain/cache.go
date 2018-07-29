package blockchain

import (
	"time"

	"github.com/hashicorp/golang-lru"
)

type cacheValue struct {
	value      interface{}
	expiration *time.Time
}

// Cache represents a container for storing objects in memory.
// Internally it uses an LRU cache allowing old items to be removed. It also
// supports an adding items with explicit expiration which are deleted lazily during
// the addition of new items.
type Cache struct {
	// container is an lru cache
	container *lru.Cache
}

// NewCache creates a new cache
func NewCache(capacity int) *Cache {
	cache := new(Cache)
	cache.container, _ = lru.New(capacity)
	return cache
}

// Add adds an item to the cache. It the cache becomes
// full, the oldest item is deleted to make room for the new value.
func (c *Cache) Add(key, val interface{}) {
	c.removeExpired()
	c.container.Add(key, &cacheValue{
		value:      val,
		expiration: nil,
	})
}

// AddWithExp adds an item to the cache with an explicit expiration time.
// Expired items are removed lazily - Whenever Add or AddWithExp are called.
// An item with an expiry time does not need to be the oldest in the cache before it is removed.
func (c *Cache) AddWithExp(key, val interface{}, expTime time.Time) {
	c.removeExpired()
	c.container.Add(key, &cacheValue{
		value:      val,
		expiration: &expTime,
	})
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
		if cVal, ok := c.container.Peek(k); ok {
			if cVal.(*cacheValue).expiration == nil {
				continue
			}
			if time.Now().After(*cVal.(*cacheValue).expiration) {
				c.container.Remove(k)
			}
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
