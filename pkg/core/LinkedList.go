package core

import (
	"bytes"
	"fmt"
	"sync"
)

var NO_CONSUMER = []byte{}

type MessageBatch struct {
	Id   []byte
	Data []byte
}

type Queue interface {
	IsEmpty() bool
	IsFull() bool
	Enqueue(id []byte, val []byte)
	BatchEnqueue(messages []MessageBatch)
	Dequeue(amount uint32, consumer []byte, time int64) ([][]byte, [][]byte)
	Unlock(consumer []byte, id []byte) bool
	UnlockAndEmpty(consumer []byte, position []byte) error
	UnlockAndEmptyBatch(consumer []byte, position [][]byte) error
	BatchUnlock(consumer []byte, position [][]byte) error
	UnlockAllFromConsumer(consumerID []byte)
	Capacity() uint32
	Size() uint32
}

// ListNode represents a node in the singly linked list.
type ListNode struct {
	ID         []byte
	Val        []byte
	Next       *ListNode
	ConsumerID []byte
	time       int64
	fetched    uint32
}

func (n *ListNode) Reset() {
	n.ID = nil
	n.Val = nil
	n.Next = nil
	n.ConsumerID = nil
	n.time = 0
	n.fetched = 0
}

// LinkedList represents a singly linked list.
type LinkedList struct {
	Head       *ListNode
	Tail       *ListNode
	Amount     Counter
	cap        uint32
	mutex      *sync.Mutex
	nodeTTL    int64
	objPool    *NodeObjectPool
	deadLetter *Queue
	fetchLimit uint32
}

// creates a queue
func NewQueue(size, fetchLimit uint32, nodeTTL int64, deadLetter *Queue) Queue {
	objPool := NewNodeObjectPool(int(size))

	return &LinkedList{
		cap: size,
		Amount: Counter{
			count: 0,
			lock:  &sync.Mutex{},
		},
		mutex:      &sync.Mutex{},
		nodeTTL:    nodeTTL,
		objPool:    objPool,
		deadLetter: deadLetter,
		fetchLimit: fetchLimit,
	}
}

// AddNode adds a new node with the given id to the linked list.
func (ll *LinkedList) Enqueue(id []byte, val []byte) {
	newNode := ll.objPool.GetObject()
	newNode.ID = id
	newNode.Val = val
	newNode.ConsumerID = NO_CONSUMER
	newNode.time = 0

	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	if ll.Amount.Value() >= int(ll.cap) {
		return
	}

	if ll.Head == nil {
		newNode.Next = ll.Head
		ll.Head = newNode
		ll.Tail = newNode
		ll.Amount.Increment()
		return
	}

	ll.Tail.Next = newNode
	ll.Tail = newNode
	ll.Amount.Increment()
}

func (ll *LinkedList) BatchEnqueue(messages []MessageBatch) {
	if len(messages) == 0 {
		return
	}

	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	if ll.Amount.Value() >= int(ll.cap) {
		return
	}

	headNode := ll.objPool.GetObject()
	headNode.ID = messages[0].Id
	headNode.Val = messages[0].Data
	headNode.ConsumerID = NO_CONSUMER
	headNode.time = 0

	currentEnd := headNode

	for i := 1; i < len(messages); i++ {
		node := ll.objPool.GetObject()
		node.ID = messages[i].Id
		node.Val = messages[i].Data
		node.ConsumerID = NO_CONSUMER
		node.time = 0

		currentEnd.Next = node
		currentEnd = node
	}

	if ll.Head == nil {
		currentEnd.Next = ll.Head
		ll.Head = headNode
		ll.Tail = currentEnd
		ll.Amount.IncrementBy(len(messages))
		return
	}

	ll.Tail.Next = headNode
	ll.Tail = currentEnd
	ll.Amount.IncrementBy(len(messages))
}

// SetConsumerID sets the ConsumerID of the ListNode with the given id to 0 if the ConsumerID matches.
func (ll *LinkedList) Unlock(consumerID []byte, id []byte) bool {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	var prev *ListNode = nil
	current := ll.Head
	for current != nil {
		if bytes.Equal(current.ID, id) {
			if bytes.Equal(current.ConsumerID, consumerID) {
				current.ConsumerID = NO_CONSUMER
				current.fetched++

				if current.fetched >= ll.fetchLimit {
					ll.handleThresholdExpire(prev, current)
					ll.objPool.ReleaseObject(current)
					ll.Amount.Decrement()
				}

				return true
			}

			break
		}

		prev = current
		current = current.Next
	}

	return false
}

func (ll *LinkedList) UnlockAndEmpty(consumerID []byte, id []byte) error {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	if ll.Head == nil {
		return fmt.Errorf("queue is empty")
	}

	return ll.unlockAndEmpty(consumerID, id)
}

func (ll *LinkedList) unlockAndEmpty(consumerID []byte, id []byte) error {
	if bytes.Equal(ll.Head.ID, id) && bytes.Equal(ll.Head.ConsumerID, consumerID) {
		if ll.Tail == ll.Head {
			ll.objPool.ReleaseObject(ll.Head)
			ll.Head = nil
			ll.Tail = nil
			ll.Amount.Decrement()
			return nil
		}

		old := ll.Head
		ll.Head = ll.Head.Next

		ll.objPool.ReleaseObject(old)

		ll.Amount.Decrement()
		return nil
	}

	prev := ll.Head
	current := ll.Head.Next

	for current != nil {
		if bytes.Equal(current.ID, id) && bytes.Equal(current.ConsumerID, consumerID) {
			if current == ll.Tail {
				ll.Tail = prev
				prev.Next = nil
			} else {
				prev.Next = current.Next
			}

			ll.objPool.ReleaseObject(current)
			ll.Amount.Decrement()
			return nil
		}

		prev = current
		current = current.Next
	}

	return fmt.Errorf("could not find item with id: %v", id)
}

func (ll *LinkedList) UnlockAndEmptyBatch(consumerID []byte, ids [][]byte) error {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	if ll.Head == nil {
		return fmt.Errorf("queue is empty")
	}

	if len(ids) == 1 {
		return ll.unlockAndEmpty(consumerID, ids[0])
	}

	if ll.Amount.Value() < len(ids) {
		return fmt.Errorf("could not all ids to ack, queue less than length of ids")
	}

	nodes := []*ListNode{}

	if bytes.Equal(ll.Head.ConsumerID, consumerID) && byteSliceExists(ids, ll.Head.ID) {
		nodes = append(nodes, ll.Head)
		ll.Head = ll.Head.Next
	}

	var prev *ListNode = nil
	current := ll.Head

	for current != nil && len(nodes) < len(ids) {
		if bytes.Equal(current.ConsumerID, consumerID) && byteSliceExists(ids, current.ID) {
			if current == ll.Tail {
				nodes = append(nodes, current)
				if ll.Head == ll.Tail {
					ll.Head = nil
					ll.Tail = nil
				} else {
					ll.Tail = prev
					ll.Tail.Next = nil
				}
				break
			}

			nodes = append(nodes, current)

			if prev == nil {
				ll.Head = ll.Head.Next
				current = ll.Head
			} else {
				prev.Next = current.Next
				current = current.Next
			}

			continue
		}

		prev = current
		current = current.Next
	}

	if len(nodes) == 0 {
		return fmt.Errorf("could not all ids to ack")
	}

	if len(nodes) != len(ids) {
		if len(nodes) == 1 {
			next := ll.Head.Next
			ll.Head.Next = nodes[0]
			nodes[0].Next = next
			return fmt.Errorf("could not all ids to ack")
		}

		current := nodes[0]
		for i := 1; i < len(nodes); i++ {
			current.Next = nodes[i]
			current = current.Next
		}

		next := ll.Head.Next
		ll.Head.Next = nodes[0]
		nodes[len(nodes)-1].Next = next

		return fmt.Errorf("could not all ids to ack")
	}

	return nil
}

type pairNode struct {
	prev    *ListNode
	current *ListNode
}

func (ll *LinkedList) BatchUnlock(consumerID []byte, ids [][]byte) error {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	if ll.Head == nil {
		return fmt.Errorf("queue is empty")
	}

	if ll.Amount.Value() < len(ids) {
		return fmt.Errorf("could not all ids to ack, queue less than length of ids")
	}

	nodes := []pairNode{}
	if bytes.Equal(ll.Head.ConsumerID, consumerID) && byteSliceExists(ids, ll.Head.ID) {
		nodes = append(nodes, pairNode{
			prev:    nil,
			current: ll.Head,
		})
	}

	current := ll.Head.Next
	prev := ll.Head
	for current != nil && len(nodes) < len(ids) {
		if bytes.Equal(current.ConsumerID, consumerID) && byteSliceExists(ids, current.ID) {
			nodes = append(nodes, pairNode{
				prev:    prev,
				current: current,
			})
		}

		current = current.Next
	}

	if len(nodes) == 0 || len(nodes) != len(ids) {
		return fmt.Errorf("could not all ids to nack")
	}

	for _, nodePair := range nodes {
		nodePair.current.ConsumerID = NO_CONSUMER
		nodePair.current.fetched++

		if nodePair.current.fetched >= ll.fetchLimit {
			ll.handleThresholdExpire(nodePair.prev, nodePair.current)
			ll.objPool.ReleaseObject(nodePair.current)
			ll.Amount.Decrement()
		}
	}

	return nil
}

func (ll *LinkedList) handleThresholdExpire(prev, current *ListNode) {
	if ll.deadLetter != nil {
		(*ll.deadLetter).Enqueue(current.ID, current.Val)
	}

	if current == ll.Tail && current == ll.Head {
		ll.Head = nil
		ll.Tail = nil
	} else if current == ll.Tail {
		ll.Tail = prev
		ll.Tail.Next = nil
	} else if current == ll.Head {
		ll.Head = current.Next
	} else {
		prev.Next = current.Next
	}
}

// GetNextNodesValWithConsumerID returns the next X number of nodes' val property that have the
// ConsumerID equal to the one passed in, or if the ConsumerID is 0.
func (ll *LinkedList) Dequeue(amount uint32, consumer []byte, time int64) ([][]byte, [][]byte) {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	data := [][]byte{}
	ids := [][]byte{}

	if ll.Head == nil {
		return data, ids
	}

	current := ll.Head

	// Traverse the list and add nodes to the result that have the correct ConsumerID.
	for current != nil && len(data) < int(amount) {
		if bytes.Equal(current.ConsumerID, consumer) || // is locked by consumer
			bytes.Equal(current.ConsumerID, NO_CONSUMER) ||
			(time-current.time) > ll.nodeTTL { // ttl has ran out

			data = append(data, current.Val)
			ids = append(ids, current.ID)
			current.ConsumerID = consumer
			current.time = time

		}
		current = current.Next
	}

	return data, ids
}

func (ll *LinkedList) UnlockAllFromConsumer(consumerID []byte) {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	current := ll.Head
	for current != nil {
		if bytes.Equal(current.ConsumerID, consumerID) {
			current.ConsumerID = NO_CONSUMER
		}
		current = current.Next
	}
}

func (ll *LinkedList) IsEmpty() bool {
	return ll.Head == nil
}

func (ll *LinkedList) Capacity() uint32 {
	return ll.cap
}

func (ll *LinkedList) IsFull() bool {
	return ll.Amount.Value() >= int(ll.cap)
}

func (ll *LinkedList) Size() uint32 {
	return uint32(ll.Amount.Value())
}

func byteSliceExists(slice [][]byte, id []byte) bool {
	for _, s := range slice {
		if bytes.Equal(s, id) {
			return true
		}
	}
	return false
}
