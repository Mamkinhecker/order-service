package cache

import (
	"order-service/internal/db"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]*db.Order
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]*db.Order),
	}
}

func (c *Cache) Set(order *db.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[order.OrderUID] = order
}

func (c *Cache) Get(uid string) (*db.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	order, exists := c.items[uid]
	return order, exists
}

func (c *Cache) GetAll() map[string]*db.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items
}
