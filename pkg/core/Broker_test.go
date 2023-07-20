package core_test

import (
	"sybline/pkg/core"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

var lgr = hclog.New(&hclog.LoggerOptions{
	Name:  "sybline-logger",
	Level: hclog.LevelFromString("debug"),
})

func TestBroker_CreateQueue_Sync(t *testing.T) {
	q := core.NewQueueManager(5 * 60)
	b := core.NewBroker(q, lgr)

	routingKey := "route"

	err := b.CreateQueue("one", routingKey, 5, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	err = b.CreateQueue("two", routingKey, 5, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to existing routing key")
}

func TestBroker_Can_Send_Via_Message_Routing(t *testing.T) {
	q := core.NewQueueManager(5 * 60)
	b := core.NewBroker(q, lgr)

	routingKey := "route"

	err := b.CreateQueue("one", routingKey, 1, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	err = b.CreateQueue("two", routingKey, 1, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to existing routing key")

	err = b.CreateQueue("three", "random", 1, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to another new routing key")

	isFull, err := q.IsFull("one")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty")

	isFull, err = q.IsFull("two")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty")

	isFull, err = q.IsFull("three")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty")

	one := uuid.New()
	err = b.AddMessage(routingKey, []byte("random"), one[:])
	assert.Nil(t, err, "should be able to add a message")

	isFull, err = q.IsFull("one")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, true, "should be full as has single message")

	isFull, err = q.IsFull("two")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, true, "should be full as has single message")

	isFull, err = q.IsFull("three")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty again")
}

func TestBroker_Delete_Queue(t *testing.T) {
	q := core.NewQueueManager(5 * 60)
	b := core.NewBroker(q, lgr)

	routingKey := "route"

	err := b.CreateQueue("one", routingKey, 1, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	one := uuid.New()
	err = b.AddMessage(routingKey, []byte("random"), one[:])
	assert.Nil(t, err, "should be able to add a message")

	err = b.DeleteQueue("one")
	assert.Nil(t, err, "should be able to delete queue")

	two := uuid.New()
	err = b.AddMessage(routingKey, []byte("random"), two[:])
	assert.Nil(t, err, "should have a silent drop")
}

func TestBroker_Add_Route_Key(t *testing.T) {
	q := core.NewQueueManager(5 * 60)
	b := core.NewBroker(q, lgr)

	routingKey := "route"
	routingKeyTwo := "routeTwo"

	err := b.CreateQueue("one", routingKey, 1, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	err = b.AddRouteKey(routingKeyTwo, "one")
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	// three as we default to 3 shards
	for i := 0; i < 3; i++ {
		id := uuid.New()
		err = b.AddMessage(routingKeyTwo, []byte("random"), id[:])
		assert.Nil(t, err, "should be able to add a message")
	}

	isFull, err := q.IsFull("one")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, true, "should be full")
}

func TestBroker_Delete_Routing_Keys(t *testing.T) {
	q := core.NewQueueManager(5 * 60)
	b := core.NewBroker(q, lgr)

	routingKey := "route"
	routingKeyTwo := "routeTwo"

	err := b.CreateQueue("one", routingKey, 1, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	err = b.AddRouteKey(routingKeyTwo, "one")
	assert.Nil(t, err, "should be able to connect new queue to new routing key: two")

	err = b.DeleteRoutingKey(routingKeyTwo, "one")
	assert.Nil(t, err, "should be able to delete the first key")

	err = b.DeleteRoutingKey(routingKey, "one")
	assert.NotNil(t, err, "should not be able to delete the second key")
}

func TestBroker_Can_Batch_Messages(t *testing.T) {
	q := core.NewQueueManager(5 * 60)
	b := core.NewBroker(q, lgr)

	routingKey := "route"

	err := b.CreateQueue("one_batch", routingKey, 3, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	err = b.CreateQueue("two_batch", routingKey, 3, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to existing routing key")

	err = b.CreateQueue("three_batch", "random", 3, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to another new routing key")

	isFull, err := q.IsFull("one_batch")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty")

	isFull, err = q.IsFull("two_batch")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty")

	isFull, err = q.IsFull("three_batch")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty")

	one := uuid.New()
	err = b.BatchAddMessage(routingKey, [][]byte{
		[]byte("random"), []byte("random"), []byte("random"),
	}, [][]byte{one[:], one[:], one[:]})
	assert.Nil(t, err, "should be able to add a message")

	isFull, err = q.IsFull("one_batch")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, true, "should be full as has single message")

	isFull, err = q.IsFull("two_batch")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, true, "should be full as has single message")

	isFull, err = q.IsFull("three_batch")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty again")
}

func TestBroker_Can_Batch_Messages_Empty_Message(t *testing.T) {
	q := core.NewQueueManager(5 * 60)
	b := core.NewBroker(q, lgr)

	routingKey := "route"

	err := b.CreateQueue("one", routingKey, 3, 3, false)
	assert.Nil(t, err, "should be able to connect new queue to new routing key")

	isFull, err := q.IsFull("one")
	assert.Nil(t, err, "no error should be produced as queue should exist")
	assert.Equal(t, isFull, false, "should be empty")

	one := uuid.New()
	err = b.BatchAddMessage(routingKey, [][]byte{
		{}, []byte("random"), []byte("random"),
	}, [][]byte{one[:], one[:], one[:]})
	assert.NotNil(t, err, "should be able to add a message")
}
