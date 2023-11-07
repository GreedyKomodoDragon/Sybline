package handler

import (
	"sync"
)

type ObjectPool[T any] struct {
	pool    chan T
	maxSize int
	mutex   sync.Mutex
	killed  bool
	factory func() T
}

func NewObjectPool[T any](maxSize int, factory func() T) *ObjectPool[T] {
	pool := make(chan T, maxSize)
	return &ObjectPool[T]{
		pool:    pool,
		maxSize: maxSize,
		killed:  false,
		factory: factory,
	}
}

func (op *ObjectPool[T]) GetObject() T {
	select {
	case obj := <-op.pool:
		return obj
	default:
		op.mutex.Lock()
		defer op.mutex.Unlock()

		return op.factory()
	}
}

func (op *ObjectPool[T]) ReleaseObject(obj T) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	if len(op.pool) < op.maxSize {
		op.pool <- obj
	}
}

func (op *ObjectPool[T]) Close() {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	// Empty the channel
	for len(op.pool) > 0 {
		<-op.pool
	}

	op.killed = true

	close(op.pool)
}
