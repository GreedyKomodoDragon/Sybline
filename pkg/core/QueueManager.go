package core

import (
	"errors"
	"fmt"
	"sync"
)

var ErrQueueMissing = errors.New("queue does not exist")
var ErrQueueLockMissing = errors.New("queue lock does not exist")
var ErrQueueAlreadyExists = errors.New("queue with name already exists")
var ErrFailedEmpty = errors.New("unable to remove")
var ErrInvalidQueueSize = errors.New("unable to create queue with size of 0")

type QueueManager interface {
	CreateQueue(name string, size uint32) error
	AddMessage(name string, data []byte, id []byte) error
	BatchAddMessage(name string, data [][]byte, ids [][]byte) error
	GetMessage(name string, amount uint32, consumerID []byte, time int64) ([]Message, error)
	IsFull(name string) (bool, error)
	Ack(name string, consumerID []byte, id []byte) error
	BatchAck(name string, consumerID []byte, id [][]byte) error
	Nack(name string, consumerID []byte, id []byte) error
	BatchNack(name string, consumerID []byte, id [][]byte) error
	Delete(name string) error
	Exists(name string) bool
	ReleaseAllLocksByConsumer(consumerID []byte)
}

type Message struct {
	Id   []byte
	Data []byte
}

type queueManager struct {
	queues      map[string]Queue
	creationMux *sync.Mutex
	nodeTTL     int64
}

func NewQueueManager(nodeTTL int64) QueueManager {
	return &queueManager{
		queues:      map[string]Queue{},
		creationMux: &sync.Mutex{},
		nodeTTL:     nodeTTL,
	}
}

func (q queueManager) CreateQueue(name string, size uint32) error {
	if q.queues == nil {
		return fmt.Errorf("queue map not initialised")
	}

	if len(name) == 0 {
		return ErrInvalidQueueName
	}

	if size == 0 {
		return ErrInvalidQueueSize
	}

	// to ensure queue cannot be deleted
	q.creationMux.Lock()
	defer q.creationMux.Unlock()

	if _, ok := q.queues[name]; ok {
		return ErrQueueAlreadyExists
	}

	q.queues[name] = NewQueue(size, q.nodeTTL)

	return nil
}

func (q queueManager) AddMessage(name string, data []byte, id []byte) error {
	queue, ok := q.queues[name]
	if !ok {
		return ErrQueueMissing
	}

	queue.Enqueue(id, data)
	return nil
}

func (q queueManager) IsFull(name string) (bool, error) {
	queue, ok := q.queues[name]
	if !ok {
		return false, ErrQueueMissing
	}

	return queue.IsFull(), nil
}

func (q queueManager) GetMessage(name string, amount uint32, consumerID []byte, time int64) ([]Message, error) {
	messages := []Message{}
	found := 0

	queue, ok := q.queues[name]
	if !ok {
		return messages, ErrQueueMissing
	}

	if !ok {
		if len(messages) > 0 {
			return messages, nil
		}
		return nil, ErrQueueLockMissing
	}

	data, ids := queue.Dequeue(amount-uint32(found), consumerID, time)

	for index := range ids {
		messages = append(messages, Message{
			Id:   ids[index],
			Data: data[index],
		})
	}

	return messages, nil
}

func (q queueManager) Ack(name string, consumerID []byte, id []byte) error {
	queue, ok := q.queues[name]
	if !ok {
		return ErrQueueMissing
	}

	return queue.UnlockAndEmpty(consumerID, id)
}

func (q queueManager) BatchAck(name string, consumerID []byte, id [][]byte) error {
	queue, ok := q.queues[name]
	if !ok {
		return ErrQueueMissing
	}

	return queue.UnlockAndEmptyBatch(consumerID, id)
}

func (q queueManager) Nack(name string, consumerID []byte, id []byte) error {
	queue, ok := q.queues[name]
	if !ok {
		return ErrQueueMissing
	}

	if !queue.Unlock(consumerID, id) {
		return ErrFailedEmpty
	}

	return nil
}

func (q queueManager) BatchNack(name string, consumerID []byte, id [][]byte) error {
	queue, ok := q.queues[name]
	if !ok {
		return ErrQueueMissing
	}

	return queue.BatchUnlock(consumerID, id)
}

func (q queueManager) Delete(name string) error {
	_, ok := q.queues[name]
	if !ok {
		return ErrQueueMissing
	}

	delete(q.queues, name)
	return nil
}

func (q queueManager) Exists(name string) bool {
	_, ok := q.queues[name]
	return ok
}

func (q queueManager) ReleaseAllLocksByConsumer(consumerID []byte) {
	for name := range q.queues {
		queue, ok := q.queues[name]
		if !ok {
			continue
		}

		queue.UnlockAllFromConsumer(consumerID)
	}
}

func (q queueManager) BatchAddMessage(name string, data [][]byte, ids [][]byte) error {
	queue, ok := q.queues[name]
	if !ok {
		return ErrQueueMissing
	}
	// create as much as possible before locking
	msgs := []MessageBatch{}
	for i := 0; i < len(data); i++ {
		msgs = append(msgs, MessageBatch{
			Id:   ids[i],
			Data: data[i],
		})
	}

	queue.BatchEnqueue(msgs)

	return nil
}
