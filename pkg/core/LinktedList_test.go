package core_test

import (
	"sybline/pkg/core"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueue(t *testing.T) {
	one := uuid.New()
	size := uint32(3)
	q := core.NewQueue(size, 3, 2*60, nil)

	// Test Enqueue for an empty queue.
	q.Enqueue(one[:], []byte("value 1"))
	assert.False(t, q.IsEmpty(), "Expected queue to not be empty after Enqueue.")

	// Test Enqueue for a non-empty queue.
	two := uuid.New()
	q.Enqueue(two[:], []byte("value 2"))
	assert.Equal(t, uint32(2), q.Size(), "Expected queue amount to be 2 after two Enqueues.")

	// Test Enqueue for a full queue.
	for i := uint32(3); i <= size; i++ {
		id := uuid.New()
		q.Enqueue(id[:], []byte("value"))
	}
	assert.True(t, q.IsFull(), "Expected queue to be full after %d Enqueues.", size)
}

func TestDequeue(t *testing.T) {
	// create a new linked list

	size := uint32(10)
	ll := core.NewQueue(size, 3, 2*60, nil).(*core.LinkedList)

	// add nodes to the linked list
	one := uuid.New()
	two := uuid.New()
	three := uuid.New()
	four := uuid.New()

	ll.Enqueue(one[:], []byte("node1"))
	ll.Enqueue(two[:], []byte("node2"))
	ll.Enqueue(three[:], []byte("node3"))
	ll.Enqueue(four[:], []byte("node4"))

	// dequeue nodes from the linked list
	data, ids := ll.Dequeue(4, []byte("a"), time.Now().Unix())

	require.Equal(t, 4, len(data), "expected 2 nodes to be dequeued, got %d", len(data))
	require.Equal(t, 4, len(ids), "expected 2 nodes to be dequeued, got %d", len(ids))

	require.Equal(t, "node1", string(data[0]), "expected nodes with values 'node1', got '%s'", string(data[0]))
	require.Equal(t, "node2", string(data[1]), "expected nodes with values 'node2', got '%s'", string(data[1]))
	require.Equal(t, "node3", string(data[2]), "expected nodes with values 'node1', got '%s'", string(data[0]))
	require.Equal(t, "node4", string(data[3]), "expected nodes with values 'node2', got '%s'", string(data[1]))

	require.Equal(t, [][]byte{one[:], two[:], three[:], four[:]}, ids, "expected ids to be [%v, %v, %v, %v], got %v", one[:], two[:], three[:], four[:], ids)

	require.Equal(t, []byte("a"), ll.Head.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.ConsumerID)
	require.Equal(t, []byte("a"), ll.Head.Next.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.Next.ConsumerID)
}

func TestBatchEnqueue(t *testing.T) {
	// create a new linked list
	size := uint32(10)
	ll := core.NewQueue(size, 3, 2*60, nil).(*core.LinkedList)

	zero := uuid.New()
	ll.Enqueue(zero[:], []byte("node9"))

	one := uuid.New()
	two := uuid.New()
	three := uuid.New()
	four := uuid.New()

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one[:],
			Data: []byte("node1"),
		},
		{
			Id:   two[:],
			Data: []byte("node2"),
		},
		{
			Id:   three[:],
			Data: []byte("node3"),
		},
		{
			Id:   four[:],
			Data: []byte("node4"),
		},
	})

	five := uuid.New()
	six := uuid.New()
	seven := uuid.New()
	eight := uuid.New()

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   five[:],
			Data: []byte("node5"),
		},
		{
			Id:   six[:],
			Data: []byte("node6"),
		},
		{
			Id:   seven[:],
			Data: []byte("node7"),
		},
		{
			Id:   eight[:],
			Data: []byte("node8"),
		},
	})

	// dequeue nodes from the linked list
	data, ids := ll.Dequeue(9, []byte("a"), time.Now().Unix())

	require.Equal(t, 9, len(data), "expected 9 nodes to be dequeued, got %d", len(data))
	require.Equal(t, 9, len(ids), "expected 9 nodes to be dequeued, got %d", len(ids))

	require.Equal(t, []byte("a"), ll.Head.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.ConsumerID)
	require.Equal(t, []byte("a"), ll.Head.Next.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.Next.ConsumerID)
}

func TestBatchEnqueue_Unlock(t *testing.T) {
	size := uint32(10)
	ll := core.NewQueue(size, 3, 2*60, nil).(*core.LinkedList)

	one := uuid.New()
	two := uuid.New()
	three := uuid.New()
	four := uuid.New()

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one[:],
			Data: []byte("node1"),
		},
		{
			Id:   two[:],
			Data: []byte("node2"),
		},
		{
			Id:   three[:],
			Data: []byte("node3"),
		},
		{
			Id:   four[:],
			Data: []byte("node4"),
		},
	})

	five := uuid.New()
	six := uuid.New()
	seven := uuid.New()
	eight := uuid.New()

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   five[:],
			Data: []byte("node5"),
		},
		{
			Id:   six[:],
			Data: []byte("node6"),
		},
		{
			Id:   seven[:],
			Data: []byte("node7"),
		},
		{
			Id:   eight[:],
			Data: []byte("node8"),
		},
	})

	// dequeue nodes from the linked list
	_, ids := ll.Dequeue(2, []byte("a"), time.Now().Unix())

	require.Equal(t, [][]byte{one[:], two[:]}, ids, "expected ids to be [%v, %v], got %v", ids[0], ids[1], ids)

	require.Equal(t, []byte("a"), ll.Head.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.ConsumerID)
	require.Equal(t, []byte("a"), ll.Head.Next.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.Next.ConsumerID)

	require.Nil(t, ll.UnlockAndEmpty([]byte("a"), one[:]), "should be nil")

	_, ids = ll.Dequeue(2, []byte("a"), time.Now().Unix())
	require.Equal(t, [][]byte{two[:], three[:]}, ids, "expected ids to be [%v, %v], got %v", ids[0], ids[1], ids)

	require.Nil(t, ll.UnlockAndEmpty([]byte("a"), three[:]), "should be nil")

	_, ids = ll.Dequeue(2, []byte("a"), time.Now().Unix())
	require.Equal(t, [][]byte{two[:], four[:]}, ids, "expected ids to be [%v, %v], got %v", ids[0], ids[1], ids)
}

func TestBatchEnqueue_UnlockEmptyBatch(t *testing.T) {
	size := uint32(10)
	ll := core.NewQueue(size, 3, 2*60, nil).(*core.LinkedList)

	one := []byte("one")
	two := []byte("two")
	three := []byte("three")
	four := []byte("four")

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one[:],
			Data: []byte("node1"),
		},
		{
			Id:   two[:],
			Data: []byte("node2"),
		},
		{
			Id:   three[:],
			Data: []byte("node3"),
		},
		{
			Id:   four[:],
			Data: []byte("node4"),
		},
	})

	five := []byte("five")
	six := []byte("six")
	seven := []byte("seven")
	eight := []byte("eight")

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   five[:],
			Data: []byte("node5"),
		},
		{
			Id:   six[:],
			Data: []byte("node6"),
		},
		{
			Id:   seven[:],
			Data: []byte("node7"),
		},
		{
			Id:   eight[:],
			Data: []byte("node8"),
		},
	})

	// dequeue nodes from the linked list
	_, ids := ll.Dequeue(2, []byte("a"), time.Now().Unix())

	require.Equal(t, [][]byte{one[:], two[:]}, ids, "expected ids to be [%v, %v], got %v", ids[0], ids[1], ids)

	require.Equal(t, []byte("a"), ll.Head.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.ConsumerID)
	require.Equal(t, []byte("a"), ll.Head.Next.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.Next.ConsumerID)

	require.Nil(t, ll.UnlockAndEmptyBatch([]byte("a"), [][]byte{one[:]}), "should be nil")

	_, ids = ll.Dequeue(3, []byte("a"), time.Now().Unix())
	require.Equal(t, [][]byte{two[:], three[:], four[:]}, ids, "expected ids to be [%v, %v, %v], got %v", ids[0], ids[1], ids[2], ids)

	require.Nil(t, ll.UnlockAndEmptyBatch([]byte("a"), [][]byte{four[:], three[:]}), "should be nil")

	_, ids = ll.Dequeue(2, []byte("a"), time.Now().Unix())
	require.Equal(t, [][]byte{two[:], five[:]}, ids, "expected ids to be [%v, %v], got %v", string(two), string(five), ids)

	require.NotNil(t, ll.UnlockAndEmptyBatch([]byte("a"), [][]byte{four[:], eight[:]}), "should return an error")

	_, ids = ll.Dequeue(2, []byte("a"), time.Now().Unix())
	require.Equal(t, [][]byte{two[:], five[:]}, ids, "expected ids to be [%v, %v], got %v", ids[0], ids[1], ids)

}

func TestBatchEnqueue_UnlockEmptyBatch_Empty_Add(t *testing.T) {
	size := uint32(10)
	ll := core.NewQueue(size, 3, 2*60, nil).(*core.LinkedList)

	one := []byte("one")
	two := []byte("two")

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one,
			Data: []byte("node1"),
		},
		{
			Id:   two[:],
			Data: []byte("node2"),
		},
	})

	// dequeue nodes from the linked list
	_, ids := ll.Dequeue(2, []byte("a"), time.Now().Unix())

	require.Nil(t, ll.UnlockAndEmptyBatch([]byte("a"), [][]byte{ids[0], ids[1]}), "should be nil")

	three := uuid.New()
	four := uuid.New()
	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   three[:],
			Data: []byte("node1"),
		},
		{
			Id:   four[:],
			Data: []byte("node2"),
		},
	})

	_, ids = ll.Dequeue(2, []byte("a"), time.Now().Unix())
	require.Equal(t, [][]byte{three[:], four[:]}, ids, "expected ids to be [%v, %v], got %v", ids[0], ids[1], ids)

	require.Nil(t, ll.UnlockAndEmptyBatch([]byte("a"), [][]byte{four[:], three[:]}), "should be nil")
}

func TestBatchEnqueue_UnlockBatch(t *testing.T) {
	size := uint32(10)
	ll := core.NewQueue(size, 3, 2*60, nil).(*core.LinkedList)

	one := []byte("one")
	two := []byte("two")
	three := []byte("three")
	four := []byte("four")

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one,
			Data: []byte("node1"),
		},
		{
			Id:   two,
			Data: []byte("node2"),
		},
		{
			Id:   three,
			Data: []byte("node3"),
		},
		{
			Id:   four,
			Data: []byte("node4"),
		},
	})

	// dequeue nodes from the linked list
	_, ids := ll.Dequeue(2, []byte("a"), time.Now().Unix())

	require.Equal(t, [][]byte{one, two}, ids, "expected ids to be [%v, %v], got %v", one[:], two[:], ids)
	require.Equal(t, []byte("a"), ll.Head.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.ConsumerID)
	require.Equal(t, []byte("a"), ll.Head.Next.ConsumerID, "expected ConsumerID of dequeued nodes to be set to 0, got %d", ll.Head.Next.ConsumerID)

	require.Nil(t, ll.BatchUnlock([]byte("a"), [][]byte{one, two}), "should be nil")

	_, ids = ll.Dequeue(3, []byte("b"), time.Now().Unix())
	require.Equal(t, [][]byte{one, two, three}, ids, "expected ids to be [%v, %v, %v], got %v", string(one), string(two), string(three), ids)
}

func TestBatchEnqueue_Unlock_Batch(t *testing.T) {
	size := uint32(10)
	ll := core.NewQueue(size, 3, 2*60, nil).(*core.LinkedList)

	one := []byte("one")
	two := []byte("two")
	three := []byte("three")
	four := []byte("four")

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one[:],
			Data: []byte("node1"),
		},
		{
			Id:   two[:],
			Data: []byte("node2"),
		},
		{
			Id:   three[:],
			Data: []byte("node3"),
		},
		{
			Id:   four[:],
			Data: []byte("node4"),
		},
	})

	consumerID := []byte("a")
	consumerIDTwo := []byte("a")

	// dequeue nodes from the linked list
	_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
	require.Equal(t, [][]byte{one, two}, ids, "expected ids to be [%v, %v], got %v", ids[0], ids[1], ids)

	require.Nil(t, ll.BatchUnlock(consumerID, [][]byte{one, two}), "should be nil")

	_, ids = ll.Dequeue(2, consumerIDTwo, time.Now().Unix())
	require.Equal(t, [][]byte{one, two}, ids, "expected ids to be [%v, %v], got %v", one, two, ids)
}
func TestBatchEnqueue_Unlock_Batch_DLQ(t *testing.T) {
	size := uint32(10)
	dlq := core.NewQueue(size, 3, 2*60, nil)

	ll := core.NewQueue(size, 3, 2*60, &dlq).(*core.LinkedList)

	one := []byte("one")
	two := []byte("two")
	three := []byte("three")
	four := []byte("four")

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one[:],
			Data: []byte("node1"),
		},
		{
			Id:   two[:],
			Data: []byte("node2"),
		},
		{
			Id:   three[:],
			Data: []byte("node3"),
		},
		{
			Id:   four[:],
			Data: []byte("node4"),
		},
	})

	consumerID := []byte("a")

	// dequeue nodes from the linked list
	for i := 0; i < 3; i++ {
		_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
		require.Equal(t, [][]byte{one, two}, ids, "expected ids to be [%v, %v], got %v", one, two, ids)
		require.Nil(t, ll.BatchUnlock(consumerID, [][]byte{one, two}), "should be nil")
	}

	_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
	require.Equal(t, [][]byte{three, four}, ids, "expected ids to be [%v, %v], got %v", three, four, ids)

	for i := 0; i < 3; i++ {
		_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
		require.Equal(t, [][]byte{three, four}, ids, "expected ids to be [%v, %v], got %v", one, two, ids)
		require.Nil(t, ll.BatchUnlock(consumerID, [][]byte{four, three}), "should be nil")
	}

	_, ids = ll.Dequeue(2, consumerID, time.Now().Unix())
	require.Equal(t, [][]byte{}, ids, "should be empty")
}

func TestBatchEnqueue_Unlock_DLQ(t *testing.T) {
	size := uint32(10)
	dlq := core.NewQueue(size, 3, 2*60, nil)

	ll := core.NewQueue(size, 3, 2*60, &dlq).(*core.LinkedList)

	one := []byte("one")
	two := []byte("two")
	three := []byte("three")
	four := []byte("four")

	ll.BatchEnqueue([]core.MessageBatch{
		{
			Id:   one[:],
			Data: []byte("node1"),
		},
		{
			Id:   two[:],
			Data: []byte("node2"),
		},
		{
			Id:   three[:],
			Data: []byte("node3"),
		},
		{
			Id:   four[:],
			Data: []byte("node4"),
		},
	})

	consumerID := []byte("a")

	// dequeue nodes from the linked list
	for i := 0; i < 3; i++ {
		_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
		require.Equal(t, [][]byte{one, two}, ids, "expected ids to be [%v, %v], got %v", one, two, ids)
		require.True(t, ll.Unlock(consumerID, one), "should be nil")
	}

	for i := 0; i < 3; i++ {
		_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
		require.Equal(t, [][]byte{two, three}, ids, "expected ids to be [%v, %v], got %v", two, three, ids)
		require.True(t, ll.Unlock(consumerID, three), "should be nil")
	}

	for i := 0; i < 3; i++ {
		_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
		require.Equal(t, [][]byte{two, four}, ids, "expected ids to be [%v, %v], got %v", two, four, ids)
		require.True(t, ll.Unlock(consumerID, four), "should be nil")
	}

	_, ids := ll.Dequeue(2, consumerID, time.Now().Unix())
	require.Equal(t, [][]byte{two}, ids, "expected ids to be [%v], got %v", two, ids)
}
