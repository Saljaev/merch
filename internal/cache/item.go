package cache

import (
	"sync"
	"time"
)

// Item is a record of type any in Cache
type Item struct {
	expires *time.Time
	data    any
	sync.RWMutex
}

func (item *Item) touch(duration time.Duration) {
	item.Lock()
	expiration := time.Now().Add(duration)
	item.expires = &expiration
	item.Unlock()
}

func (item *Item) expired() bool {
	var value bool

	item.RLock()

	if item.expires == nil {
		value = true
	} else {
		value = item.expires.Before(time.Now())
	}

	item.RUnlock()
	return value
}
