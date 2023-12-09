package fsm

import (
	"fmt"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/rbac"
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
	rbacManager    rbac.RoleManager
}

func NewSyblineFSM(broker core.Broker, consumer core.ConsumerManager, auth auth.AuthManager, queueManager core.QueueManager, rbacManager rbac.RoleManager) (syblineFSM, error) {
	return syblineFSM{
		broker:         broker,
		consumer:       consumer,
		auth:           auth,
		queueManager:   queueManager,
		commandPayload: &CommandPayload{},
		batchMessage:   &structs.BatchMessages{},
		rbacManager:    rbacManager,
	}, nil
}

// Apply log is invoked once a log entry is committed.
func (b syblineFSM) Apply(lg raft.Log) (interface{}, error) {
	// ignore if RAFT_LOG
	if lg.LogType == raft.RAFT_LOG || len(lg.Data) == 0 {
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
		var payCasted structs.QueueInfo
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_CREATE_QUEUE)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.broker.CreateQueue(payCasted.Name, payCasted.RoutingKey, payCasted.Size,
			payCasted.RetryLimit, payCasted.HasDLQueue)

	case SUBMIT_MESSAGE:
		var payCasted structs.MessageInfo
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasPermission(payload.Username, payCasted.Rk, rbac.SUBMIT_MESSAGE_ACTION)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.broker.AddMessage(payCasted.Rk, payCasted.Data, payCasted.Id)

	case SUBMIT_BATCH_MESSAGE:
		payCasted := b.batchMessage
		defer payCasted.Reset()

		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		for _, data := range payCasted.Messages {
			ok, err := b.rbacManager.HasPermission(payload.Username, data.RK, rbac.SUBMIT_BATCH_ACTION)
			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, fmt.Errorf("does not have permission to perform action")
			}
		}

		var errIn error = nil
		for _, data := range payCasted.Messages {
			if err := b.broker.BatchAddMessage(data.RK, data.Data, data.Ids); err != nil {
				errIn = err
			}
		}

		return nil, errIn

	case ADD_ROUTING_KEY:
		var payCasted structs.AddRoute
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_ADD_ROUTING_KEY)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.broker.AddRouteKey(payCasted.RouteName, payCasted.QueueName)

	case DELETE_ROUTING_KEY:
		var payCasted structs.DeleteRoute
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_DELETE_ROUTING_KEY)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.broker.DeleteRoutingKey(payCasted.RouteName, payCasted.QueueName)

	case GET_MESSAGES:
		var payCasted structs.RequestMessageData
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasPermission(payload.Username, payCasted.QueueName, rbac.GET_MESSAGES_ACTION)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		messages, err := b.consumer.GetMessages(payCasted.QueueName, payCasted.Amount, payCasted.ConsumerID, payCasted.Time)
		if err != nil {
			return nil, err
		}

		return messages, nil

	case ACK:
		var payCasted structs.AckUpdate
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasPermission(payload.Username, payCasted.QueueName, rbac.ACK_ACTION)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.consumer.Ack(payCasted.QueueName, payCasted.Id, payCasted.ConsumerID)

	case NACK:
		var payCasted structs.AckUpdate
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasPermission(payload.Username, payCasted.QueueName, rbac.ACK_ACTION)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.consumer.Nack(payCasted.QueueName, payCasted.Id, payCasted.ConsumerID)

	case DELETE_QUEUE:
		var payCasted structs.DeleteQueueInfo
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_DELETE_QUEUE)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.broker.DeleteQueue(payCasted.QueueName)

	case CREATE_ACCOUNT:
		var payCasted structs.UserCreds
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_CREATE_USER)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.auth.CreateUser(payCasted.Username, payCasted.Password)

	case CHANGE_PASSWORD:
		var payCasted structs.ChangeCredentials
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_CHANGE_PASSWORD)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return b.auth.ChangePassword(payCasted.Username, payCasted.OldPassword, payCasted.NewPassword)

	case DELETE_USER:
		var payCasted structs.UserInformation
		if err := msgpack.Unmarshal(payload.Data, &payCasted); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_DELETE_USER)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.auth.DeleteUser(payCasted.Username)

	case BATCH_ACK:
		var data structs.BatchAckUpdate
		if err := msgpack.Unmarshal(payload.Data, &data); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasPermission(payload.Username, data.QueueName, rbac.BATCH_ACK_ACTION)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.consumer.BatchAck(data.QueueName, data.Ids, data.ConsumerID)

	case BATCH_NACK:
		var data structs.BatchNackUpdate
		if err := msgpack.Unmarshal(payload.Data, &data); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasPermission(payload.Username, data.QueueName, rbac.BATCH_ACK_ACTION)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.consumer.BatchNack(data.QueueName, data.Ids, data.ConsumerID)

	case CREATE_ROLE:
		var data rbac.Role
		if err := msgpack.Unmarshal(payload.Data, &data); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_CREATE_ROLE)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		return nil, b.rbacManager.AddRole(data)

	case ASSIGN_ROLE:
		var data structs.RoleUsername
		if err := msgpack.Unmarshal(payload.Data, &data); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_ASSIGN_ROLE)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		if !b.auth.UserExists(data.Username) {
			return nil, fmt.Errorf("user with name '%s' does not exist", data.Username)
		}

		if !b.rbacManager.RoleExists(data.Role) {
			return nil, fmt.Errorf("role with name '%s' does not exist", data.Role)
		}

		return nil, b.rbacManager.AssignRole(data.Username, data.Role)

	case UNASSIGN_ROLE:
		var data structs.RoleUsername
		if err := msgpack.Unmarshal(payload.Data, &data); err != nil {
			return nil, err
		}

		ok, err := b.rbacManager.HasAdminPermission(payload.Username, rbac.ALLOW_UNASSIGN_ROLE)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, fmt.Errorf("does not have permission to perform action")
		}

		if !b.auth.UserExists(data.Username) {
			return nil, fmt.Errorf("user with name '%s' does not exist", data.Username)
		}

		if !b.rbacManager.RoleExists(data.Role) {
			return nil, fmt.Errorf("role with name '%s' does not exist", data.Role)
		}

		return nil, b.rbacManager.UnassignRole(data.Username, data.Role)
	}

	log.Error().Uint32("action", uint32(payload.Op)).Msg("not raft log command type")
	return nil, nil
}
