package cache

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {

	var cache *Cache

	BeforeEach(func() {
		cache = NewCache(10)
	})

	Describe(".Add", func() {
		It("should successfully add an item", func() {
			key := "myKey"
			val := "myVal"
			Expect(cache.container.Len()).To(Equal(0))
			cache.Add(key, val)
			Expect(cache.container.Len()).To(Equal(1))
		})
	})

	Describe(".AddWithExp", func() {
		It("should successfully add an item", func() {
			key := "myKey"
			val := "myVal"
			Expect(cache.container.Len()).To(Equal(0))
			cache.AddWithExp(key, val, time.Now().Add(time.Hour))
			Expect(cache.container.Len()).To(Equal(1))
		})

		It("should successfully add an item and remove expired items", func() {
			cache.AddWithExp("some_key", "some_value", time.Now().Add(time.Millisecond*100))
			time.Sleep(100 * time.Millisecond)
			Expect(cache.container.Len()).To(Equal(1))
			cache.AddWithExp("key", "val", time.Now().Add(time.Hour))
			Expect(cache.container.Len()).To(Equal(1))
		})

		It("should successfully add an item and not remove unexpired items", func() {
			cache.AddWithExp("some_key", "some_value", time.Now().Add(time.Millisecond*100))
			time.Sleep(10 * time.Millisecond)
			Expect(cache.container.Len()).To(Equal(1))
			cache.AddWithExp("key", "val", time.Now().Add(time.Hour))
			Expect(cache.container.Len()).To(Equal(2))
		})
	})

	Describe(".NewActiveCache", func() {

		Context("cache should remove expired item", func() {
			CacheItemRemovalInterval = 50 * time.Millisecond
			cache := NewActiveCache(1)

			It("should successfully remove item", func() {
				cache.AddWithExp("key1", "value1", time.Now().Add(10*time.Millisecond))
				Expect(cache.Len()).To(Equal(1))
				time.Sleep(100 * time.Millisecond)
				Expect(cache.Len()).To(Equal(0))
			})
		})
	})

	Describe(".AddMulti", func() {

		var cache *Cache
		BeforeEach(func() {
			cache = NewCache(10)
		})

		It("should successfully add multiple values", func() {
			values := []interface{}{1, 2, "3"}
			cache.AddMulti(time.Time{}, values)
			Expect(cache.Len()).To(Equal(1))
		})
	})

	Describe(".HasMulti", func() {

		var cache *Cache
		var values []interface{}

		Context("when multi-value serialized key exist in the cache", func() {
			BeforeEach(func() {
				cache = NewCache(10)
				values = []interface{}{1, 2, "3"}
				cache.AddMulti(time.Time{}, values...)
				Expect(cache.Len()).To(Equal(1))
			})

			It("should return true", func() {
				has := cache.HasMulti(values...)
				Expect(has).To(BeTrue())
			})
		})

		Context("when multi-value serialized key does not exist in cache", func() {
			BeforeEach(func() {
				cache = NewCache(10)
				values = []interface{}{1, 2, "3"}
				cache.AddMulti(time.Time{}, values...)
				Expect(cache.Len()).To(Equal(1))
			})

			It("should when multi-value serialized key exist in the cache", func() {
				has := cache.HasMulti([]interface{}{1, 2, 3})
				Expect(has).To(BeFalse())
			})
		})
	})

	Describe(".Peek", func() {
		It("should return value of item", func() {
			cache.Add("some_key", "some_value")
			val := cache.Peek("some_key")
			Expect(val).To(Equal("some_value"))
		})

		It("should return nil if item does not exist", func() {
			val := cache.Peek("some_key")
			Expect(val).To(BeNil())
		})
	})

	Describe(".Get", func() {
		It("should return value of item", func() {
			cache.Add("some_key", "some_value")
			val := cache.Get("some_key")
			Expect(val).To(Equal("some_value"))
		})

		It("should return nil if item does not exist", func() {
			val := cache.Get("some_key")
			Expect(val).To(BeNil())
		})
	})

	Describe(".Has", func() {
		It("should return true if item exists", func() {
			cache.Add("k1", "some_value")
			Expect(cache.Has("k1")).To(BeTrue())
		})

		It("should return false if item does not exists", func() {
			cache.Add("k1", "some_value")
			Expect(cache.Has("k2")).To(BeFalse())
		})
	})

	Describe(".Keys", func() {
		It("should return two keys (k1, k2)", func() {
			cache.Add("k1", "some_value")
			cache.Add("k2", "some_value2")
			Expect(cache.Keys()).To(HaveLen(2))
			Expect(cache.Keys()).To(Equal([]interface{}{"k1", "k2"}))
		})

		It("should return empty", func() {
			keys := cache.Keys()
			Expect(keys).To(HaveLen(0))
			Expect(keys).To(Equal([]interface{}{}))
		})
	})

	Describe(".Remove", func() {
		It("should successfully remove item", func() {
			cache.Add("k1", "some_value")
			cache.Add("k2", "some_value2")
			cache.Remove("k1")
			Expect(cache.Has("k1")).To(BeFalse())
			Expect(cache.Has("k2")).To(BeTrue())
		})
	})

	Describe(".Len", func() {
		It("should successfully return length = 2", func() {
			cache.Add("k1", "some_value")
			cache.Add("k2", "some_value2")
			Expect(cache.Len()).To(Equal(2))
		})
	})

})
