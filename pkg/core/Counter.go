package core

import (
	"sync"
)

type Counter struct {
	count int
	lock  *sync.Mutex
}

func (c *Counter) Increment() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count++
}

func (c *Counter) IncrementBy(amount int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count += amount
}

func (c *Counter) Decrement() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count -= 1
}

func (c *Counter) DecrementBy(amount int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.count -= amount
}

func (c *Counter) Value() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.count
}
