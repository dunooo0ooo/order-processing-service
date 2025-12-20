package order_cache

import (
	"sync"
	"time"

	"github.com/dunooo0ooo/wb-tech-l0/internal/entity"
)

type OrderCache interface {
	Get(id string) (*entity.Order, bool)
	Set(id string, o *entity.Order)
}

type OrderCacheImpl struct {
	mu    sync.RWMutex
	data  map[string]entry
	limit int
}

type entry struct {
	o       entity.Order
	touched time.Time
}

func NewOrderCache(limit int) *OrderCacheImpl {
	if limit <= 0 {
		limit = 1000
	}
	return &OrderCacheImpl{
		data:  make(map[string]entry, limit),
		limit: limit,
	}
}

func (c *OrderCacheImpl) Get(id string) (*entity.Order, bool) {
	c.mu.RLock()
	e, ok := c.data[id]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	c.mu.Lock()
	if e2, ok2 := c.data[id]; ok2 {
		e2.touched = time.Now()
		c.data[id] = e2
	}
	c.mu.Unlock()

	o := e.o
	return &o, true
}

func (c *OrderCacheImpl) Set(id string, o *entity.Order) {
	if id == "" || o == nil {
		return
	}

	c.mu.Lock()
	c.data[id] = entry{o: *o, touched: time.Now()}
	c.evictIfNeededLocked()
	c.mu.Unlock()
}

func (c *OrderCacheImpl) evictIfNeededLocked() {
	if len(c.data) <= c.limit {
		return
	}

	var (
		oldestK string
		oldestT time.Time
		first   = true
	)

	for k, v := range c.data {
		if first || v.touched.Before(oldestT) {
			oldestK, oldestT = k, v.touched
			first = false
		}
	}

	if oldestK != "" {
		delete(c.data, oldestK)
	}
}
