package structs

import (
	"sync"
)

// ObjectPool is a generic object pool implementation in Go.
type ObjectPool[T any] struct {
	pool    chan T     // Channel for storing objects
	maxSize int        // Maximum size of the pool
	mutex   sync.Mutex // Mutex for synchronizing access to the pool
	killed  bool       // Flag indicating whether the pool is closed
	factory func() T   // Factory function to create new objects
}

// NewObjectPool creates a new object pool with the specified maximum size and factory function.
func NewObjectPool[T any](maxSize int, factory func() T) *ObjectPool[T] {
	pool := make(chan T, maxSize)
	return &ObjectPool[T]{
		pool:    pool,
		maxSize: maxSize,
		killed:  false,
		factory: factory,
	}
}

// GetObject retrieves an object from the pool. If the pool is empty, it creates a new object using the factory function.
func (op *ObjectPool[T]) GetObject() T {
	select {
	case obj := <-op.pool:
		return obj
	default:
		op.mutex.Lock()
		defer op.mutex.Unlock()

		// Create a new object using the factory function
		return op.factory()
	}
}

// ReleaseObject releases an object back to the pool if the pool is not full.
func (op *ObjectPool[T]) ReleaseObject(obj T) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	// If the pool is not full, release the object back to the pool
	if len(op.pool) < op.maxSize {
		op.pool <- obj
	}
}

// Close closes the object pool by emptying the channel and setting the killed flag.
func (op *ObjectPool[T]) Close() {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	// Empty the channel by discarding all objects
	for len(op.pool) > 0 {
		<-op.pool
	}

	// Set the killed flag to true
	op.killed = true

	// Close the channel
	close(op.pool)
}
