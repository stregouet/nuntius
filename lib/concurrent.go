package lib

import (
	"sync"
)

type ConcurrentList struct {
	mu      sync.RWMutex
	content []interface{}
}

func NewConcurrentList(c []interface{}) *ConcurrentList {
	return &ConcurrentList{
		content: c,
	}
}

func (c *ConcurrentList) Length() int {
	return len(c.content)
}

func (c *ConcurrentList) Last() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	l := c.Length()
	if l > 0 {
		return c.content[l-1]
	}
	return nil
}

func (c *ConcurrentList) Push(item interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.content = append(c.content, item)
}

func (c *ConcurrentList) Remove(toremove interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, item := range c.content {
		if item == toremove {
			c.content = append(c.content[:i], c.content[i+1:]...)
			break
		}
	}
}
