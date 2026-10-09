package pager

import "github.com/suhaasya/wisp-go/internal/driver"

type cacheKey struct {
	offset int
	limit  int
}

type windowCache struct {
	max   int
	order []cacheKey
	items map[cacheKey]*driver.Window
}

func newWindowCache(max int) windowCache {
	return windowCache{
		max:   max,
		items: make(map[cacheKey]*driver.Window),
	}
}

func (c *windowCache) len() int {
	return len(c.items)
}

func (c *windowCache) get(offset, limit int) (*driver.Window, bool) {
	k := cacheKey{offset: offset, limit: limit}
	w, ok := c.items[k]
	if !ok {
		return nil, false
	}
	for i, key := range c.order {
		if key == k {
			copy(c.order[i:], c.order[i+1:])
			c.order = c.order[:len(c.order)-1]
			break
		}
	}
	c.order = append(c.order, k)
	return w, true
}

func (c *windowCache) put(offset, limit int, w *driver.Window) {
	k := cacheKey{offset: offset, limit: limit}
	if _, ok := c.items[k]; ok {
		c.get(offset, limit)
	} else if len(c.order) >= c.max && c.max > 0 {
		ev := c.order[0]
		c.order = c.order[1:]
		delete(c.items, ev)
	}
	c.items[k] = w
	c.order = append(c.order, k)
}
