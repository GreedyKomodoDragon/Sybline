package handler

import (
	"sybline/pkg/structs"
	"sync"
)

type GetObjectPool struct {
	pool    chan *structs.RequestMessageData
	maxSize int
	mutex   sync.Mutex
	killed  bool
}

func NewGetObjectPool(maxSize int) *GetObjectPool {
	pool := make(chan *structs.RequestMessageData, maxSize)
	return &GetObjectPool{
		pool:    pool,
		maxSize: maxSize,
		killed:  false,
	}
}

func (op *GetObjectPool) GetObject() *structs.RequestMessageData {
	select {
	case obj := <-op.pool:
		return obj
	default:
		op.mutex.Lock()
		defer op.mutex.Unlock()

		if op.killed {
			return nil
		}

		if len(op.pool) >= op.maxSize {
			// create but we won't accept it back
			return &structs.RequestMessageData{}
		}

		obj := &structs.RequestMessageData{}
		return obj
	}
}

func (op *GetObjectPool) ReleaseObject(obj *structs.RequestMessageData) {
	obj.Reset()

	op.mutex.Lock()
	defer op.mutex.Unlock()

	if len(op.pool) < op.maxSize {
		op.pool <- obj
	}
}

func (op *GetObjectPool) Close() {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	// empty the channel
	for len(op.pool) > 0 {
		<-op.pool
	}

	op.killed = true

	close(op.pool)
}
