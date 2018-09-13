package histcache

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"

	"github.com/hashicorp/golang-lru"
)

// MultiKey represents a key consisting of a slice of objects
type MultiKey []interface{}

// HistoryCache serves the purpose of remembering interactions between
// peers for the avoidance of duplicate work like sending the same
// transaction or other items multiple times to a node. This package
// helps to reduce feedback loop and spam.
type HistoryCache struct {
	cache *lru.Cache
}

// NewHistoryCache creates a new history cache
func NewHistoryCache(size int) (*HistoryCache, error) {
	hc := new(HistoryCache)
	c, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	hc.cache = c
	return hc, nil
}

// Add a slice of objects. The items can mean whatever the caller intends
func (hc *HistoryCache) Add(h MultiKey) error {

	bs, err := asn1.Marshal(h)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(bs)
	hc.cache.Add(hex.EncodeToString(hash[:]), 0x1)
	return nil
}

// Has checks if a slice of
func (hc *HistoryCache) Has(h MultiKey) bool {

	bs, err := asn1.Marshal(h)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(bs)
	_, yes := hc.cache.Get(hex.EncodeToString(hash[:]))
	return yes
}

// Len returns the number of items in the cache
func (hc *HistoryCache) Len() int {
	return hc.cache.Len()
}

// Keys returns the keys
func (hc *HistoryCache) Keys() []interface{} {
	return hc.cache.Keys()
}
