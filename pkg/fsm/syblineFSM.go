package fsm

import (
	"fmt"
	"os"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/structs"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
)

var BREAK_SYMBOL = []byte("§")

type syblineFSM struct {
	broker         core.Broker
	consumer       core.ConsumerManager
	queueManager   core.QueueManager
	auth           auth.AuthManager
	commandPayload *CommandPayload
	batchMessage   *structs.BatchMessages
}

func NewSyblineFSM(broker core.Broker, consumer core.ConsumerManager, auth auth.AuthManager, queueManager core.QueueManager) (syblineFSM, error) {
	return syblineFSM{
		broker:         broker,
		consumer:       consumer,
		auth:           auth,
		queueManager:   queueManager,
		commandPayload: &CommandPayload{},
		batchMessage:   &structs.BatchMessages{},
	}, nil
}

// Apply log is invoked once a log entry is committed.
func (b syblineFSM) Apply(lg raft.Log) (interface{}, error) {
	// ignore if RAFT_LOG
	if lg.LogType == raft.RAFT_LOG || lg.LogType == 0 {
		return nil, nil
	}

	// re-use object
	payload := b.commandPayload
	defer payload.Reset()

	if err := msgpack.Unmarshal(lg.Data, &payload); err != nil {
		log.Error().Uint64("logType", lg.LogType).Bytes("data", lg.Data).Uint64("index", lg.Index).Err(err).Msg("issue marshalling store payload")
		return nil, err
	}

	switch payload.Op {
	case CREATE_QUEUE:
		payCasted := payload.Data.(structs.QueueInfo)
		return nil, b.broker.CreateQueue(payCasted.Name, payCasted.RoutingKey, payCasted.Size,
			payCasted.RetryLimit, payCasted.HasDLQueue)

	case SUBMIT_MESSAGE:
		payCasted := payload.Data.(structs.MessageInfo)
		return nil, b.broker.AddMessage(payCasted.Rk, payCasted.Data, payCasted.Id)

	case SUBMIT_BATCH_MESSAGE:
		payCasted := payload.Data.(structs.BatchMessages)

		var errIn error = nil
		for _, data := range payCasted.Messages {
			if err := b.broker.BatchAddMessage(data.RK, data.Data, data.Ids); err != nil {
				errIn = err
			}
		}

		return nil, errIn

	case ADD_ROUTING_KEY:
		payCasted := payload.Data.(structs.AddRoute)

		return nil, b.broker.AddRouteKey(payCasted.RouteName, payCasted.QueueName)

	case DELETE_ROUTING_KEY:
		payCasted := payload.Data.(structs.DeleteRoute)
		return nil, b.broker.DeleteRoutingKey(payCasted.RouteName, payCasted.QueueName)

	case GET_MESSAGES:
		payCasted := payload.Data.(structs.RequestMessageData)

		messages, err := b.consumer.GetMessages(payCasted.QueueName, payCasted.Amount, payCasted.ConsumerID, payCasted.Time)
		if err != nil {
			return nil, err
		}

		return messages, nil

	case ACK:
		payCasted := payload.Data.(structs.AckUpdate)
		return nil, b.consumer.Ack(payCasted.QueueName, payCasted.Id, payCasted.ConsumerID)

	case NACK:
		payCasted := payload.Data.(structs.AckUpdate)
		return nil, b.consumer.Nack(payCasted.QueueName, payCasted.Id, payCasted.ConsumerID)

	case DELETE_QUEUE:
		payCasted := payload.Data.(structs.DeleteQueueInfo)
		return nil, b.broker.DeleteQueue(payCasted.QueueName)

	case CREATE_ACCOUNT:
		payCasted := payload.Data.(structs.UserCreds)
		return nil, b.auth.CreateUser(payCasted.Username, payCasted.Password)

	case CHANGE_PASSWORD:
		payCasted := payload.Data.(structs.ChangeCredentials)
		return b.auth.ChangePassword(payCasted.Username, payCasted.OldPassword, payCasted.NewPassword)

	case REMOVE_LOCKS:
		payCasted := payload.Data.(structs.RemoveLocks)
		b.queueManager.ReleaseAllLocksByConsumer(payCasted.ConsumerID)
		return nil, nil

	case DELETE_USER:
		payCasted := payload.Data.(structs.UserInformation)

		return nil, b.auth.DeleteUser(payCasted.Username)

	case BATCH_ACK:
		data := payload.Data.(structs.BatchAckUpdate)
		return nil, b.consumer.BatchAck(data.QueueName, data.Ids, data.ConsumerID)

	case BATCH_NACK:
		data := payload.Data.(structs.BatchNackUpdate)
		return nil, b.consumer.BatchNack(data.QueueName, data.Ids, data.ConsumerID)
	}

	_, _ = fmt.Fprintf(os.Stderr, "not raft log command type\n")
	return nil, nil
}
