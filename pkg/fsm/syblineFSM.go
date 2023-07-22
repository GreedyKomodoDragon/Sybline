package fsm

import (
	"fmt"
	"io"
	"os"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/structs"
	"sybline/pkg/utils"

	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack/v5"
)

var BREAK_SYMBOL = []byte("§")

type syblineFSM struct {
	broker       core.Broker
	consumer     core.ConsumerManager
	queueManager core.QueueManager
	auth         auth.AuthManager
	store        *SyblineStore
}

func NewSyblineFSM(broker core.Broker, consumer core.ConsumerManager, auth auth.AuthManager, queueManager core.QueueManager, store *SyblineStore) (syblineFSM, error) {
	return syblineFSM{
		broker:       broker,
		consumer:     consumer,
		auth:         auth,
		queueManager: queueManager,
		store:        store,
	}, nil
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (b syblineFSM) Apply(log *raft.Log) interface{} {
	switch log.Type {
	case raft.LogCommand:
		payload := CommandPayload{}
		if err := msgpack.Unmarshal(log.Data, &payload); err != nil {
			fmt.Fprintf(os.Stderr, "error marshalling store payload %s\n", err.Error())
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}
		}

		switch payload.Op {
		case CREATE_QUEUE:
			var payCasted structs.QueueInfo
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.broker.CreateQueue(payCasted.Name, payCasted.RoutingKey, payCasted.Size,
				payCasted.RetryLimit, payCasted.HasDLQueue)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case SUBMIT_MESSAGE:
			var payCasted structs.MessageInfo
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.broker.AddMessage(payCasted.Rk, payCasted.Data, payCasted.Id)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case SUBMIT_BATCH_MESSAGE:
			var payCasted structs.BatchMessages
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			var errIn error = nil
			for _, data := range payCasted.Messages {
				if err := b.broker.BatchAddMessage(data.RK, data.Data, data.Ids); err != nil {
					errIn = err
				}
			}

			return &ApplyResponse{
				Error: errIn,
				Data:  nil,
			}

		case ADD_ROUTING_KEY:
			var payCasted structs.AddRoute
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.broker.AddRouteKey(payCasted.RouteName, payCasted.QueueName)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case DELETE_ROUTING_KEY:
			var payCasted structs.DeleteRoute
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.broker.DeleteRoutingKey(payCasted.RouteName, payCasted.QueueName)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case GET_MESSAGES:
			var payCasted structs.RequestMessageData
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			messages, err := b.consumer.GetMessages(payCasted.QueueName, payCasted.Amount, payCasted.ConsumerID, payCasted.Time)
			if err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			return &ApplyResponse{
				Error: nil,
				Data:  messages,
			}

		case ACK:
			var payCasted structs.AckUpdate
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.consumer.Ack(payCasted.QueueName, payCasted.Id, payCasted.ConsumerID)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case NACK:
			var payCasted structs.AckUpdate
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.consumer.Nack(payCasted.QueueName, payCasted.Id, payCasted.ConsumerID)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case DELETE_QUEUE:
			var payCasted structs.DeleteQueueInfo
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.broker.DeleteQueue(payCasted.QueueName)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case CREATE_ACCOUNT:
			var payCasted structs.UserCreds
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.auth.CreateUser(payCasted.Username, payCasted.Password)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case CHANGE_PASSWORD:
			var payCasted structs.ChangeCredentials
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			ok, err := b.auth.ChangePassword(payCasted.Username, payCasted.OldPassword, payCasted.NewPassword)
			return &ApplyResponse{
				Error: err,
				Data:  ok,
			}

		case REMOVE_LOCKS:
			var payCasted structs.RemoveLocks
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			b.queueManager.ReleaseAllLocksByConsumer(payCasted.ConsumerID)
			return &ApplyResponse{
				Error: nil,
				Data:  nil,
			}

		case DELETE_USER:
			var payCasted structs.UserInformation
			if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.auth.DeleteUser(payCasted.Username)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case BATCH_ACK:
			var data structs.BatchAckUpdate
			if err := msgpack.Unmarshal(payload.Data, &data); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.consumer.BatchAck(data.QueueName, data.Ids, data.ConsumerID)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}

		case BATCH_NACK:
			var data structs.BatchNackUpdate
			if err := msgpack.Unmarshal(payload.Data, &data); err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  nil,
				}
			}

			err := b.consumer.BatchNack(data.QueueName, data.Ids, data.ConsumerID)
			return &ApplyResponse{
				Error: err,
				Data:  nil,
			}
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "not raft log command type\n")
	return nil
}

// Copies logs into snapshot
func (b syblineFSM) Snapshot() (raft.FSMSnapshot, error) {
	data, err := b.getSnapshotData()
	if err != nil {
		return nil, err
	}

	return newSnapshotSybline(data)
}

func (b syblineFSM) getSnapshotData() ([]byte, error) {
	b.store.LogMux.Lock()
	defer b.store.LogMux.Unlock()

	data := []byte{}

	lastIndex, err := b.store.LastIndex()
	if err != nil {
		return nil, err
	}

	// index 0 is meta data
	log := &raft.Log{}
	for i := uint64(1); i <= lastIndex; i++ {
		if err := b.store.GetLog(i, log); err != nil {
			return nil, err
		}

		logData, err := msgpack.Marshal(log)
		if err != nil {
			return nil, err
		}

		data = append(data, logData...)
		data = append(data, BREAK_SYMBOL...)
	}

	return data, nil
}

// Restore pulls in the snapshot file and attempts to rebuild logs and state via logs
func (b syblineFSM) Restore(rClose io.ReadCloser) error {
	var bts []byte
	buffer := make([]byte, 1024)

	for {
		n, err := rClose.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		bts = append(bts, buffer[:n]...)

		// TODO: find a better way that does not require two traverisals
		indexes := utils.FindIndexes(bts, BREAK_SYMBOL)
		if len(indexes) == 0 {
			continue
		}

		previous := 0

		for i := 0; i < len(indexes); i++ {
			subSlice := bts[previous:indexes[i]]
			if len(subSlice) == 0 {
				continue
			}

			if err := b.extractApplyLog(&subSlice); err == nil {
				previous = indexes[i] + len(BREAK_SYMBOL)
			}
		}

		bts = bts[previous:]
	}

	return nil
}

// Takes bytes and then applies log to fsm if a command log
func (b syblineFSM) extractApplyLog(by *[]byte) error {
	payload := raft.Log{}
	if err := msgpack.Unmarshal(*by, &payload); err != nil {
		return err
	}

	if err := b.store.StoreLog(&payload); err != nil {
		return err
	}

	if payload.Type != raft.LogCommand {
		return nil
	}

	b.Apply(&payload)
	return nil
}
