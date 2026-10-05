package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}
type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	mu       sync.Mutex
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{capacity: capacity, items: make(map[K]*list.Element)}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.items[key]
	if !ok {
		return
	}
	value = v.Value.(entry[K, V]).value
	c.ll.MoveToFront(v)

	return
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.capacity <= 0 {
		return
	}

	if v, ok := c.items[key]; ok {
		v.Value = entry[K, V]{key: key, value: value}
		return
	}

	if len(c.items) >= c.capacity {
		oldest := c.ll.Back()
		if oldest != nil {
			c.ll.Remove(oldest)
			delete(c.items, oldest.Value.(entry[K, V]).key)
		}
	}

	newEntry := entry[K, V]{key: key, value: value}
	v := c.ll.PushFront(newEntry)
	c.items[key] = v
}
