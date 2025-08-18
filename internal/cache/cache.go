package cache

import (
	"fmt"
	"wb/internal/models"
)

type Cache struct {
	cacheMap map[string]models.Order
	keys     []string
	size     int
}

func NewCache(size int) *Cache {
	return &Cache{
		cacheMap: make(map[string]models.Order),
		keys:     make([]string, 0, size),
		size:     size,
	}
}

func (c *Cache) PutInCache(key string, order models.Order) {
	if _, ok := c.cacheMap[key]; ok {
		c.cacheMap[key] = order
		return
	}

	if len(c.keys) >= c.size {
		oldest := c.keys[0]
		c.keys = c.keys[1:]
		delete(c.cacheMap, oldest)
	}

	c.cacheMap[key] = order
	c.keys = append(c.keys, key)
}

func (c *Cache) GetIfInCache(key string) (models.Order, bool) {
	order, ok := c.cacheMap[key]
	return order, ok
}

func (c *Cache) Show() {
	fmt.Println(c.keys)
}
