package handler

import (
	"context"
	"errors"
	"fmt"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/fsm"
	"sybline/pkg/rbac"
	"sybline/pkg/structs"
	"time"

	raft "github.com/GreedyKomodoDragon/raft"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack"
)

var (
	ErrNoAuthToken        = errors.New("no valid auth token provided")
	ErrInvalidCredentials = errors.New("invalid credentials could not login")
	ErrNotLeader          = errors.New("cannot communicate with non-leader nodes")
	ErrRaftNotIntialised  = errors.New("raft server not running")
	ErrResponseCastFailed = errors.New("unable to cast response to the expected type")
	ErrMetaDataNotFound   = errors.New("unable to get authenication metadata")
	EMPTY_MESSAGES        = []*MessageData{}
)

type MessageData struct {
	Id   []byte `json:"id,omitempty"`
	Data []byte `json:"data,omitempty"`
}

type MessageSubmit struct {
	RoutingKey string `json:"routingKey,omitempty"`
	Data       []byte `json:"data,omitempty"`
}

type Handler interface {
	GetMessages(ctx context.Context, queue string, amount uint32) ([]*MessageData, error)
	SubmitMessage(ctx context.Context, routingKey string, data []byte) error
	CreateQueue(ctx context.Context, routingKey string, name string, size uint32, retryLimit uint32, hasDLQueue bool) error
	Ack(ctx context.Context, queue string, id []byte) error
	Login(ctx context.Context, username string, password string) (string, error)
	ChangePassword(ctx context.Context, username string, oldPass string, newPass string) error
	Nack(ctx context.Context, queue string, id []byte) error
	DeleteQueue(ctx context.Context, queue string) error
	AddRoutingKey(ctx context.Context, routingKey string, queue string) error
	DeleteRoutingKey(ctx context.Context, routingKey string, queue string) error
	CreateUser(ctx context.Context, username string, password string) error
	SubmitBatchedMessages(ctx context.Context, messages []MessageSubmit) error
	IsLeaderNode(ctx context.Context) (bool, error)
	DeleteUser(ctx context.Context, username string) error
	BatchAck(ctx context.Context, queue string, ids [][]byte) error
	BatchNack(ctx context.Context, queue string, ids [][]byte) error
	LogOut(ctx context.Context) error
	CreateRole(ctx context.Context, role string) error
	AssignRole(ctx context.Context, role string, username string) error
	UnassignRole(ctx context.Context, role string, username string) error
}

type handler struct {
	rbacManager    rbac.RoleManager
	authManager    auth.AuthManager
	raftServer     raft.Raft
	salt           string
	getObjectPool  *ObjectPool[structs.RequestMessageData]
	getCommandPool *ObjectPool[fsm.CommandPayload]
}

func NewHandler(rbacManager rbac.RoleManager, authManager auth.AuthManager, raftServer raft.Raft, salt string) Handler {
	return &handler{
		rbacManager: rbacManager,
		authManager: authManager,
		raftServer:  raftServer,
		salt:        salt,
		getObjectPool: NewObjectPool[structs.RequestMessageData](10000, func() structs.RequestMessageData {
			return structs.RequestMessageData{}
		}),
		getCommandPool: NewObjectPool[fsm.CommandPayload](10000, func() fsm.CommandPayload {
			return fsm.CommandPayload{}
		}),
	}
}

func (h *handler) AddRoutingKey(ctx context.Context, routingKey string, queue string) error {
	if err := validateRoutingKeyChange(queue, routingKey); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_ADD_ROUTING_KEY)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.ADD_ROUTING_KEY, structs.AddRoute{
		RouteName: routingKey,
		QueueName: queue,
	}, username)

	return err
}

func (h *handler) SubmitMessage(ctx context.Context, routingKey string, data []byte) error {
	if err := validateMessage(data, routingKey); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasPermission(username, routingKey, rbac.SUBMIT_MESSAGE_ACTION)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	id := uuid.New()
	_, err = h.sendCommand(fsm.SUBMIT_MESSAGE, structs.MessageInfo{
		Rk:   routingKey,
		Data: data,
		Id:   id[:],
	}, username)

	return err
}

func (h *handler) GetMessages(ctx context.Context, queue string, amount uint32) ([]*MessageData, error) {
	if err := validateRequestMessages(queue, amount); err != nil {
		return EMPTY_MESSAGES, err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return EMPTY_MESSAGES, err
	}

	consumerID, err := getBytesFromCtx(ctx, "consumerID")
	if err != nil {
		return EMPTY_MESSAGES, err
	}

	ok, err := h.rbacManager.HasPermission(username, queue, rbac.GET_MESSAGES_ACTION)
	if err != nil {
		return EMPTY_MESSAGES, err
	}

	if !ok {
		return EMPTY_MESSAGES, fmt.Errorf("does not have permission to perform action")
	}

	msg := h.getObjectPool.GetObject()
	msg.QueueName = queue
	msg.Amount = amount
	msg.ConsumerID = consumerID
	msg.Time = time.Now().Unix()

	res, err := h.sendCommand(fsm.GET_MESSAGES, msg, username)

	h.getObjectPool.ReleaseObject(msg)
	if err != nil {
		return EMPTY_MESSAGES, err
	}

	msgs, _ := res.([]core.Message)
	msgConverted := make([]*MessageData, len(msgs))

	for index, msg := range msgs {
		msgConverted[index] = &MessageData{
			Id:   msg.Id,
			Data: msg.Data,
		}
	}

	return msgConverted, nil
}

func (h *handler) CreateQueue(ctx context.Context, routingKey string, name string, size uint32, retryLimit uint32, hasDLQueue bool) error {
	if err := validateCreateQueue(routingKey, name, size); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_CREATE_QUEUE)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.CREATE_QUEUE, structs.QueueInfo{
		RoutingKey: routingKey,
		Name:       name,
		Size:       size,
		RetryLimit: retryLimit,
		HasDLQueue: hasDLQueue,
	}, username)

	return err
}

func (h *handler) Ack(ctx context.Context, queue string, id []byte) error {
	if err := validateAck(queue, id); err != nil {
		return err
	}

	if _, err := uuid.FromBytes(id); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	consumerID, err := getBytesFromCtx(ctx, "consumerID")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasPermission(username, queue, rbac.ACK_ACTION)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.ACK, structs.AckUpdate{
		QueueName:  queue,
		Id:         id,
		ConsumerID: consumerID,
	}, username)

	return err
}

func (h *handler) Login(ctx context.Context, username string, password string) (string, error) {
	if len(username) == 0 {
		return "", ErrInvalidUsername
	}

	if len(password) == 0 {
		return "", ErrInvalidPassword
	}

	token, err := h.authManager.Login(username, auth.GenerateHash(password, h.salt))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (h *handler) ChangePassword(ctx context.Context, username string, oldPass string, newPass string) error {
	if err := validateChangePassword(username, oldPass, newPass); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_CHANGE_PASSWORD)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.CHANGE_PASSWORD, structs.ChangeCredentials{
		Username:    username,
		OldPassword: auth.GenerateHash(oldPass, h.salt),
		NewPassword: auth.GenerateHash(newPass, h.salt),
	}, username)

	return err
}

func (h *handler) Nack(ctx context.Context, queue string, id []byte) error {
	if err := validateAck(queue, id); err != nil {
		return err
	}

	if _, err := uuid.FromBytes(id); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	consumerID, err := getBytesFromCtx(ctx, "consumerID")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasPermission(username, queue, rbac.ACK_ACTION)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.NACK, structs.AckUpdate{
		QueueName:  queue,
		Id:         id,
		ConsumerID: consumerID,
	}, username)

	return err
}

func (h *handler) DeleteQueue(ctx context.Context, queue string) error {
	if len(queue) == 0 {
		return ErrInvalidQueueName
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_DELETE_QUEUE)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.DELETE_QUEUE, structs.DeleteQueueInfo{
		QueueName: queue,
	}, username)

	return err
}

func (h *handler) DeleteRoutingKey(ctx context.Context, routingKey string, queue string) error {
	if err := validateRoutingKeyChange(queue, routingKey); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_DELETE_ROUTING_KEY)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.DELETE_ROUTING_KEY, structs.DeleteRoute{
		RouteName: routingKey,
		QueueName: queue,
	}, username)

	return err
}

func (h *handler) CreateUser(ctx context.Context, reqUsername string, password string) error {
	if err := validateUsernamePassword(reqUsername, password); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_CREATE_USER)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.CREATE_ACCOUNT, structs.UserCreds{
		Username: reqUsername,
		Password: auth.GenerateHash(password, h.salt),
	}, username)

	return err
}

type prebatch struct {
	data [][]byte
	id   [][]byte
}

func (h *handler) SubmitBatchedMessages(ctx context.Context, messages []MessageSubmit) error {
	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	routes := make(map[string]*prebatch)
	for _, msg := range messages {
		if err := validateMessage(msg.Data, msg.RoutingKey); err != nil {
			return err
		}

		if _, ok := routes[msg.RoutingKey]; ok {
			id := uuid.New()
			idRaw := id[:]

			routes[msg.RoutingKey].data = append(routes[msg.RoutingKey].data, msg.Data)
			routes[msg.RoutingKey].id = append(routes[msg.RoutingKey].id, idRaw)

			continue
		}

		// only check it the first time we see the routing key
		ok, err := h.rbacManager.HasPermission(username, msg.RoutingKey, rbac.SUBMIT_BATCH_ACTION)
		if err != nil {
			return err
		}

		if !ok {
			return fmt.Errorf("does not have permission to perform action")
		}

		id := uuid.New()
		idRaw := id[:]

		routes[msg.RoutingKey] = &prebatch{
			data: [][]byte{msg.Data},
			id:   [][]byte{idRaw},
		}
	}

	msgs := []structs.Messages{}
	for rk, data := range routes {
		msgs = append(msgs, structs.Messages{
			RK:   rk,
			Data: data.data,
			Ids:  data.id,
		})
	}

	_, err = h.sendCommand(fsm.SUBMIT_BATCH_MESSAGE, structs.BatchMessages{
		Messages: msgs,
	}, username)

	return err
}

func (h *handler) IsLeaderNode(ctx context.Context) (bool, error) {
	if h.raftServer == nil {
		return false, ErrRaftNotIntialised
	}

	return h.raftServer.State() == raft.LEADER, nil
}

func (h *handler) DeleteUser(ctx context.Context, reqUsername string) error {
	if len(reqUsername) == 0 {
		return ErrInvalidUsername
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_DELETE_USER)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.DELETE_USER, structs.UserInformation{
		Username: reqUsername,
	}, username)

	return err
}

func (h *handler) BatchAck(ctx context.Context, queue string, ids [][]byte) error {
	if err := validateBatchAck(queue, ids); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasPermission(username, queue, rbac.BATCH_ACK_ACTION)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	consumerID, err := getBytesFromCtx(ctx, "consumerID")
	if err != nil {
		return err
	}

	_, err = h.sendCommand(fsm.BATCH_ACK, structs.BatchAckUpdate{
		QueueName:  queue,
		Ids:        ids,
		ConsumerID: consumerID,
	}, username)

	return err
}

func (h *handler) BatchNack(ctx context.Context, queue string, ids [][]byte) error {
	if err := validateBatchAck(queue, ids); err != nil {
		return err
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasPermission(username, queue, rbac.BATCH_ACK_ACTION)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	consumerID, err := getBytesFromCtx(ctx, "consumerID")
	if err != nil {
		return err
	}

	_, err = h.sendCommand(fsm.BATCH_ACK, structs.BatchNackUpdate{
		QueueName:  queue,
		Ids:        ids,
		ConsumerID: consumerID,
	}, username)

	return err
}

func (h *handler) LogOut(ctx context.Context) error {
	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	token, err := getStringFromCtx(ctx, "token")
	if err != nil {
		return err
	}

	return h.authManager.LogOut(username, token)
}

func (h *handler) CreateRole(ctx context.Context, role string) error {
	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_CREATE_ROLE)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	roleConvert, err := h.rbacManager.CreateRole(role)
	if err != nil {
		return err
	}

	_, err = h.sendCommand(fsm.CREATE_ROLE, roleConvert, username)
	return err
}

func (h *handler) AssignRole(ctx context.Context, role string, reqUsername string) error {
	if len(role) == 0 || len(reqUsername) == 0 {
		return fmt.Errorf("invalid request payload for AssignRole")
	}

	if !h.authManager.UserExists(reqUsername) {
		return fmt.Errorf("user with name '%s' does not exist", reqUsername)
	}

	if !h.rbacManager.RoleExists(role) {
		return fmt.Errorf("role with name '%s' does not exist", role)
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_ASSIGN_ROLE)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.ASSIGN_ROLE, structs.RoleUsername{
		Role:     role,
		Username: reqUsername,
	}, username)

	return err
}

func (h *handler) UnassignRole(ctx context.Context, role string, reqUsername string) error {
	if len(role) == 0 || len(reqUsername) == 0 {
		return fmt.Errorf("invalid request payload for unassignRole")
	}

	if !h.authManager.UserExists(reqUsername) {
		return fmt.Errorf("user with name '%s' does not exist", reqUsername)
	}

	if !h.rbacManager.RoleExists(role) {
		return fmt.Errorf("role with name '%s' does not exist", role)
	}

	username, err := getStringFromCtx(ctx, "username")
	if err != nil {
		return err
	}

	ok, err := h.rbacManager.HasAdminPermission(username, rbac.ALLOW_ASSIGN_ROLE)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("does not have permission to perform action")
	}

	_, err = h.sendCommand(fsm.UNASSIGN_ROLE, structs.RoleUsername{
		Role:     role,
		Username: reqUsername,
	}, username)

	return err
}

func (h *handler) sendCommand(payloadType fsm.Operation, payload interface{}, username string) (interface{}, error) {
	jsonBytes, err := msgpack.Marshal(payload)
	if err != nil {
		return nil, err
	}

	obj := h.getCommandPool.GetObject()
	obj.Op = payloadType
	obj.Data = jsonBytes
	obj.Username = username

	data, err := msgpack.Marshal(obj)
	if err != nil {
		return nil, err
	}

	result, err := h.raftServer.ApplyLog(&data, raft.DATA_LOG)
	if err != nil {
		return nil, err
	}

	sybResult, ok := result.(*fsm.SyblineFSMResult)
	if !ok {
		return nil, fmt.Errorf("unable to convert to sybResult")
	}

	return sybResult.Data, sybResult.Err
}

func getStringFromCtx(ctx context.Context, key string) (string, error) {
	user := ctx.Value(key)
	if user == nil {
		return "", fmt.Errorf("user cannot be found")
	}

	str, ok := user.(string)
	if !ok {
		return "", fmt.Errorf("user cannot be found")
	}

	return str, nil
}

func getBytesFromCtx(ctx context.Context, key string) ([]byte, error) {
	user := ctx.Value(key)
	if user == nil {
		return nil, fmt.Errorf("user cannot be found")
	}

	bytes, ok := user.([]byte)
	if !ok {
		return nil, fmt.Errorf("user cannot be found")
	}

	return bytes, nil
}
