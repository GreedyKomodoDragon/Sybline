package core_test

import (
	"bytes"
	"math/rand"
	"sybline/pkg/core"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func contains(s []core.Message, e core.Message) bool {
	for _, a := range s {
		if bytes.Equal(a.Data, e.Data) && bytes.Equal(a.Id, e.Id) {
			return true
		}
	}
	return false
}

func TestQueueManager_Add_Message_To_Non_Existant_Queue(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	one := uuid.New()
	err := q.AddMessage(name, []byte("random"), one[:])
	expectedErrorMsg := core.ErrQueueMissing.Error()

	assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
}

func TestQueueManager_Add_Message_To_Existant_Queue(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	err := q.CreateQueue(name, 5, 3, false)
	assert.Nil(t, err, "should be able to create a queue")

	one := uuid.New()
	err = q.AddMessage(name, []byte("random"), one[:])
	assert.Nil(t, err, "should be able to add message to queue")
}

func TestQueueManager_Create_Duplicate_Queue_Sync(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	err := q.CreateQueue(name, 5, 3, false)
	assert.Nil(t, err, "should be able to create a queue")

	err = q.CreateQueue(name, 5, 3, false)
	errMsg := core.ErrQueueAlreadyExists.Error()
	assert.EqualErrorf(t, err, errMsg, "should not be able to create the queue again")
}

func TestQueueManager_Create_Duplicate_Queue_Async(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	var wg sync.WaitGroup
	var n atomic.Uint32

	createQueue := func(name string, size uint32) {
		r := rand.Intn(50)
		time.Sleep(time.Duration(r) * time.Microsecond)

		if err := q.CreateQueue(name, size, 3, false); err != nil {
			n.Add(1)
		}
		wg.Done()
	}

	wg.Add(3)
	go createQueue(name, 5)
	go createQueue(name, 5)
	go createQueue(name, 5)

	wg.Wait()

	assert.Equal(t, uint32(2), n.Load(), "should have two errors")

	one := uuid.New()
	err := q.AddMessage(name, []byte("random"), one[:])
	assert.Nil(t, err, "should be able to add message to queue")
}

func TestQueueManager_GetMessages(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	err := q.CreateQueue(name, 5, 3, false)
	assert.Nil(t, err, "should be able to create a queue")

	one := uuid.New()
	q.AddMessage(name, []byte("random"), one[:])
	q.AddMessage(name, []byte("random1"), one[:])
	q.AddMessage(name, []byte("random2"), one[:])
	q.AddMessage(name, []byte("random3"), one[:])
	q.AddMessage(name, []byte("random4"), one[:])

	messagesOne, err := q.GetMessage(name, 2, []byte("a"), time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Len(t, messagesOne, 2)

	messagesTwo, err := q.GetMessage(name, 2, []byte("a"), time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Len(t, messagesTwo, 2)
	assert.True(t, contains(messagesOne, messagesTwo[0]), "due to locks should get second message 0")
	assert.True(t, contains(messagesOne, messagesTwo[1]), "due to locks should get second message 1")

	messagesThree, err := q.GetMessage(name, 2, []byte("b"), time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Len(t, messagesThree, 2)

	messagesFour, err := q.GetMessage(name, 2, []byte("c"), time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Len(t, messagesFour, 1, "should be only one message left to be locked")
}

func TestQueueManager_Ack(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	zero := uuid.New()
	id := []byte("a")
	err := q.Ack(name, id, zero[:])
	expectedErrorMsg := core.ErrQueueMissing.Error()
	assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)

	err = q.CreateQueue(name, 5, 3, false)
	assert.Nil(t, err, "should be able to create a queue")

	one := uuid.New()
	q.AddMessage(name, []byte("random"), one[:])
	q.AddMessage(name, []byte("random1"), one[:])
	q.AddMessage(name, []byte("random2"), one[:])
	q.AddMessage(name, []byte("random3"), one[:])
	q.AddMessage(name, []byte("random4"), one[:])

	messages, err := q.GetMessage(name, 2, id, time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Len(t, messages, 2, "should have a length of two")

	err = q.Ack(name, id, messages[0].Id)
	assert.Nil(t, err, "should be able to ack a message")

	messagesTwo, err := q.GetMessage(name, 2, id, time.Now().Unix())
	assert.Len(t, messagesTwo, 2, "should have a length of two on messagesTwo")
	assert.Nil(t, err, "should be able to get messages from queue")

	matches := 0
	for _, message := range messagesTwo {
		if contains(messages, message) {
			matches += 1
		}
	}

	assert.Equal(t, 1, matches, "should only be one message that is the same")
}

func TestQueueManager_Nack_Existing_Message(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	one := uuid.New()
	id := []byte("a")

	err := q.Ack(name, id, one[:])
	expectedErrorMsg := core.ErrQueueMissing.Error()
	assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)

	err = q.CreateQueue(name, 5, 3, false)
	assert.Nil(t, err, "should be able to create a queue")

	q.AddMessage(name, []byte("random"), one[:])

	messages, err := q.GetMessage(name, 1, id, time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Equal(t, []byte("random"), messages[0].Data, "should be able to get message")

	err = q.Nack(name, id, messages[0].Id)
	assert.Nil(t, err, "should be able to nack a message")

	idNew := []byte("b")
	messages, err = q.GetMessage(name, 1, idNew, time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Equal(t, []byte("random"), messages[0].Data, "message should be the same")
}

func TestQueueManager_Delete_Queue(t *testing.T) {
	q := core.NewQueueManager(2 * 60)
	name := "queue_name"

	one := uuid.New()
	id := []byte("a")

	err := q.Ack(name, id, one[:])
	expectedErrorMsg := core.ErrQueueMissing.Error()
	assert.EqualErrorf(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)

	err = q.CreateQueue(name, 5, 3, false)
	assert.Nil(t, err, "should be able to create a queue")

	q.AddMessage(name, []byte("random"), one[:])

	messages, err := q.GetMessage(name, 1, id, time.Now().Unix())
	assert.Nil(t, err, "should be able to get messages from queue")
	assert.Equal(t, []byte("random"), messages[0].Data, "should be able to get message")

	err = q.Delete(name)
	assert.Nil(t, err, "should be able to delete queue")

	_, err = q.GetMessage(name, 1, id, time.Now().Unix())
	assert.NotNil(t, err, "should not be able to get messages from queue")
}
