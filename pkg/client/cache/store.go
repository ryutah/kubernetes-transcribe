package cache

import (
	"sync"

	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

// Store is a generic object storage interface. Reflector knows how to watch a server
// and update a store. A generic store is provided, which allows Reflector to be used
// as a local caching system, and an LRU store, which allows Reflector to work like a
// queue of items yet to be processed.
type Store interface {
	Add(id string, obj interface{})
	Update(id string, obj interface{})
	Delete(id string)
	List() []interface{}
	Contains() util.StringSet
	Get(id string) (item interface{}, exists bool)

	Replace(idToObj map[string]interface{})
}

type cache struct {
	lock  sync.RWMutex
	items map[string]interface{}
}

// Add inserts an item into the cache.
func (c *cache) Add(id string, obj interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[id] = obj
}

// Update sets an item in the cache to its updated state.
func (c *cache) Update(id string, obj interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[id] = obj
}

// Delete removes an item from the cache.
func (c *cache) Delete(id string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.items, id)
}

// List returns a list of all the items.
// List is completely threadsafe as long as you treat all items as immutable.
func (c *cache) List() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	list := make([]interface{}, 0, len(c.items))
	for _, item := range c.items {
		list = append(list, item)
	}
	return list
}

// Contains returns a util.StringSet containing all IDs of stored the items.
// This is a snapshot of a moment in time, and one should keep in mind that
// other go routines can add or remove items after you call this.
func (c *cache) Contains() util.StringSet {
	c.lock.RLock()
	defer c.lock.RUnlock()
	set := util.StringSet{}
	for id := range c.items {
		set.Insert(id)
	}
	return set
}

// Get returns the requested item, or sets exists=false.
// Get is completely threadsafe as long as you treat all items an immutable.
func (c *cache) Get(id string) (item interface{}, exists bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, exists = c.items[id]
	return item, exists
}

// Replace will delete the contents of 'c', using instead the given map.
// 'c' takes ownership of the map, you should not reference the map again
// after calling this function.
func (c *cache) Replace(idToObj map[string]interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items = idToObj
}

// NewStore returns a Store implemented simply with a map and a lock.
func NewStore() Store {
	return &cache{items: map[string]interface{}{}}
}
