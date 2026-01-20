package order_cache

import (
	"container/list"
	"sync"

	"github.com/dunooo0ooo/wb-tech-l0/internal/entity"
	"github.com/dunooo0ooo/wb-tech-l0/pkg/config"
)

type OrderCache interface {
	Get(id string) (*entity.Order, bool)
	Set(id string, o *entity.Order)
}

type OrderCacheImpl struct {
	mu    sync.Mutex
	limit int

	ll *list.List
	m  map[string]*list.Element
}

type lruEntry struct {
	key   string
	order entity.Order
}

func NewOrderCache(cfg config.CacheConfig) *OrderCacheImpl {
	limit := cfg.Limit
	if limit <= 0 {
		limit = 1000
	}
	return &OrderCacheImpl{
		limit: limit,
		ll:    list.New(),
		m:     make(map[string]*list.Element, limit),
	}
}

func (c *OrderCacheImpl) Get(id string) (*entity.Order, bool) {
	if id == "" {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.m[id]; ok {
		c.ll.MoveToFront(el)

		ent := el.Value.(lruEntry)
		o := ent.order
		return &o, true
	}

	return nil, false
}

func (c *OrderCacheImpl) Set(id string, o *entity.Order) {
	if id == "" || o == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.m[id]; ok {
		el.Value = lruEntry{key: id, order: *o}
		c.ll.MoveToFront(el)
		return
	}

	el := c.ll.PushFront(lruEntry{key: id, order: *o})
	c.m[id] = el

	if c.ll.Len() > c.limit {
		back := c.ll.Back()
		if back != nil {
			ent := back.Value.(lruEntry)
			delete(c.m, ent.key)
			c.ll.Remove(back)
		}
	}
}
