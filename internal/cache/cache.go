package cache

import (
	"log"
	"order-service/internal/db"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	items map[string]*db.Order
	allID []string
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]*db.Order),
		allID: make([]string, 0),
	}
}

func (c *Cache) Set(order *db.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[order.OrderUID]; exists {
		c.removeFromSlice(order.OrderUID)
	}

	c.items[order.OrderUID] = order
	c.allID = append(c.allID, order.OrderUID)
	if len(c.allID) > 10 {
		oldestID := c.allID[0]
		delete(c.items, oldestID)
		c.allID = c.allID[1:]
		log.Println("successfully deleted order from cache")
	}
}

func (c *Cache) Get(uid string) (*db.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	order, exists := c.items[uid]
	return order, exists
}

func (c *Cache) Restore(orders []db.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*db.Order)
	c.allID = make([]string, 0)

	for i := range orders {
		order := orders[i]
		c.items[order.OrderUID] = &order
		c.allID = append(c.allID, order.OrderUID)
		if len(c.allID) > 10 {
			oldestID := c.allID[0]
			delete(c.items, oldestID)
			c.allID = c.allID[1:]
			log.Println("successfully deleted order from cache during restore")
		}
	}
}

func (c *Cache) GetAll() map[string]*db.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.items
}

func (c *Cache) removeFromSlice(uid string) {
	for i, id := range c.allID {
		if id == uid {
			c.allID = append(c.allID[:i], c.allID[i+1:]...)
			break
		}
	}
}
