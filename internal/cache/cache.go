package cache

import (
	"sync"
	"time"
)

// Cache is generics map of Item, with ttl to delete and auto increment
// tll of item, which is touched
type Cache[T comparable] struct {
	mutex sync.RWMutex
	ttl   time.Duration
	items map[T]*Item
}

func (cache *Cache[T]) Set(key T, data any) {
	cache.mutex.Lock()
	item := &Item{data: data}
	item.touch(cache.ttl)
	cache.items[key] = item
	cache.mutex.Unlock()
}

// Get return Item and increase ttl of it
func (cache *Cache[T]) Get(key T) (any, bool) {
	var data any
	var found bool

	cache.mutex.Lock()

	item, exists := cache.items[key]
	if !exists || item.expired() {
		data = ""
		found = false
	} else {
		item.touch(cache.ttl)
		data = item.data
		found = true
	}

	cache.mutex.Unlock()

	return data, found
}

func (cache *Cache[T]) cleanup() {
	cache.mutex.Lock()

	for key, item := range cache.items {
		if item.expired() {
			delete(cache.items, key)
		}
	}

	cache.mutex.Unlock()
}

func (cache *Cache[T]) startCleanupTimer() {
	duration := cache.ttl

	if duration < time.Second {
		duration = time.Second
	}

	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
			case <-ticker:
				cache.cleanup()
			}
		}
	})()
}

func NewCache[T comparable](duration time.Duration) *Cache[T] {
	cache := &Cache[T]{
		ttl:   duration,
		items: make(map[T]*Item),
	}

	cache.startCleanupTimer()
	
	return cache
}
