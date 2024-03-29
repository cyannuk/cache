package main

const cacheTmpl = `package {{.Package}}

import (
	"sync"
	"time"
	{{.Import}}
)

{{$itemType := join (.TypeName | uncapitalize) "Item"}}
{{$cacheType := join (.TypeName | capitalize) "Cache"}}

type {{$itemType}} struct {
	value  {{.ValueType}}
	expiry int64
}

// isExpired checks if the cache item has expired.
func (item {{$itemType}}) isExpired() bool {
	return time.Now().UnixNano() > item.expiry
}

type {{$cacheType}} struct {
	items map[{{.KeyType}}]{{$itemType}} // The map storing cache items.
	mtx   sync.RWMutex  // Mutex for controlling concurrent access to the cache.
}

func New{{$cacheType}}() *{{$cacheType}} {
	cache := &{{$cacheType}}{
		items: make(map[{{.KeyType}}]{{$itemType}}),
	}

	go func() {
		for range time.Tick(5 * time.Second) {
			cache.mtx.Lock()

			// Iterate over the cache items and delete expired ones.
			for key, item := range cache.items {
				if item.isExpired() {
					delete(cache.items, key)
				}
			}

			cache.mtx.Unlock()
		}
	}()

	return cache
}

// Set adds a new item to the cache with the specified key, value, and time-to-live (TTL).
func (cache *{{$cacheType}}) Set(key {{.KeyType}}, value {{.ValueType}}, ttl time.Duration) {
	cache.mtx.Lock()
	cache.items[key] = {{$itemType}}{
		value:  value,
		expiry: time.Now().UnixNano() + int64(ttl),
	}
	cache.mtx.Unlock()
}

// Get retrieves the value associated with the given key from the cache.
func (cache *{{$cacheType}}) Get(key {{.KeyType}}) ({{.ValueType}}, bool) {
	cache.mtx.RLock()
	defer cache.mtx.RUnlock()

	item, found := cache.items[key]
	if !found {
		// If the key is not found, return the zero value for {{.ValueType}} and false.
		return item.value, false
	}

	if item.isExpired() {
		// If the item has expired, remove it from the cache and return the
		// value and false.
		delete(cache.items, key)
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}

// Remove removes the item with the specified key from the cache.
func (cache *{{$cacheType}}) Remove(key {{.KeyType}}) {
	cache.mtx.Lock()
	// Delete the item with the given key from the cache.
	delete(cache.items, key)
	cache.mtx.Unlock()
}

// Pop removes and returns the item with the specified key from the cache.
func (cache *{{$cacheType}}) Pop(key {{.KeyType}}) ({{.ValueType}}, bool) {
	cache.mtx.Lock()
	defer cache.mtx.Unlock()

	item, found := cache.items[key]
	if !found {
		// If the key is not found, return the zero value for {{.ValueType}} and false.
		return item.value, false
	}

	// If the key is found, delete the item from the cache.
	delete(cache.items, key)

	if item.isExpired() {
		// If the item has expired, return the value and false.
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}`
