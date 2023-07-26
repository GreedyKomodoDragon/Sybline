package core

import (
	"sync"
)

type NodeObjectPool struct {
	pool    chan *ListNode
	maxSize int
	mutex   sync.Mutex
	killed  bool
}

func NewNodeObjectPool(maxSize int) *NodeObjectPool {
	pool := make(chan *ListNode, maxSize)
	return &NodeObjectPool{
		pool:    pool,
		maxSize: maxSize,
		killed:  false,
	}
}

func (op *NodeObjectPool) GetObject() *ListNode {
	select {
	case obj := <-op.pool:
		return obj
	default:
		op.mutex.Lock()
		defer op.mutex.Unlock()

		if op.killed {
			return nil
		}

		return &ListNode{}
	}
}

func (op *NodeObjectPool) ReleaseObject(obj *ListNode) {
	obj.Reset()

	op.mutex.Lock()
	defer op.mutex.Unlock()

	if len(op.pool) < op.maxSize {
		op.pool <- obj
	}
}

func (op *NodeObjectPool) Close() {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	// empty the channel
	for len(op.pool) > 0 {
		<-op.pool
	}

	op.killed = true

	close(op.pool)
}
