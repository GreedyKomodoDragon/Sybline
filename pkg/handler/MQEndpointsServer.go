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
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoAuthToken        = errors.New("no valid auth token provided")
	ErrInvalidCredentials = errors.New("invalid credentials could not login")
	ErrNotLeader          = errors.New("cannot communicate with non-leader nodes")
	ErrRaftNotIntialised  = errors.New("raft server not running")
	ErrResponseCastFailed = errors.New("unable to cast response to the expected type")
	ErrMetaDataNotFound   = errors.New("unable to get authenication metadata")

	falseStatus = &Status{
		Status: false,
	}
)

type mQEndpointsServer struct {
	rbacManager    rbac.RoleManager
	authManager    auth.AuthManager
	raftServer     raft.Raft
	salt           string
	getObjectPool  *ObjectPool[structs.RequestMessageData]
	getCommandPool *ObjectPool[fsm.CommandPayload]
	UnimplementedMQEndpointsServer
}

func NewServer(rbacManager rbac.RoleManager, authManager auth.AuthManager, raftServer raft.Raft, salt string) MQEndpointsServer {
	return mQEndpointsServer{
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

func (s mQEndpointsServer) SubmitMessage(ctx context.Context, info *MessageInfo) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateMessage(info.Data, info.RoutingKey); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasPermission(username, info.RoutingKey, rbac.SUBMIT_MESSAGE_ACTION)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	id := uuid.New()
	_, err = s.sendCommand(fsm.SUBMIT_MESSAGE, structs.MessageInfo{
		Rk:   info.RoutingKey,
		Data: info.Data,
		Id:   id[:],
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) GetMessages(ctx context.Context, request *RequestMessageData) (*MessageCollection, error) {
	if s.raftServer.State() != raft.LEADER {
		return &MessageCollection{}, ErrNotLeader
	}

	if err := validateRequestMessages(request.QueueName, request.Amount); err != nil {
		return &MessageCollection{}, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &MessageCollection{}, ErrMetaDataNotFound
	}

	consumerID, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return &MessageCollection{}, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return &MessageCollection{}, err
	}

	ok, err = s.rbacManager.HasPermission(username, request.QueueName, rbac.GET_MESSAGES_ACTION)
	if err != nil {
		return &MessageCollection{}, err
	}

	if !ok {
		return &MessageCollection{}, fmt.Errorf("does not have permission to perform action")
	}

	msg := s.getObjectPool.GetObject()
	msg.QueueName = request.QueueName
	msg.Amount = request.Amount
	msg.ConsumerID = consumerID
	msg.Time = time.Now().Unix()

	res, err := s.sendCommand(fsm.GET_MESSAGES, msg, username)

	s.getObjectPool.ReleaseObject(msg)
	if err != nil {
		return &MessageCollection{}, err
	}

	msgs, _ := res.([]core.Message)
	msgConverted := make([]*MessageData, len(msgs))

	for index, msg := range msgs {
		msgConverted[index] = &MessageData{
			Id:   msg.Id,
			Data: msg.Data,
		}
	}

	return &MessageCollection{
		Messages: msgConverted,
	}, nil
}

func (s mQEndpointsServer) CreateQueue(ctx context.Context, request *QueueInfo) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateCreateQueue(request.RoutingKey, request.Name, request.Size); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		log.Err(err).Msg("unable to get consumer id")
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get consumer username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_CREATE_QUEUE)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.CREATE_QUEUE, structs.QueueInfo{
		RoutingKey: request.RoutingKey,
		Name:       request.Name,
		Size:       request.Size,
		RetryLimit: request.RetryLimit,
		HasDLQueue: request.HasDLQueue,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Ack(ctx context.Context, request *AckUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateAck(request.QueueName, request.Id); err != nil {
		return falseStatus, err
	}

	if _, err := uuid.FromBytes(request.Id); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	consumerID, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasPermission(username, request.QueueName, rbac.ACK_ACTION)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.ACK, structs.AckUpdate{
		QueueName:  request.QueueName,
		Id:         request.Id,
		ConsumerID: consumerID,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Login(ctx context.Context, request *Credentials) (*Status, error) {
	if len(request.Username) == 0 {
		return falseStatus, ErrInvalidUsername
	}

	if len(request.Password) == 0 {
		return falseStatus, ErrInvalidPassword
	}

	token, err := s.authManager.Login(request.Username, auth.GenerateHash(request.Password, s.salt))
	if err != nil {
		log.Err(err).Msg("unable to log in")
		return falseStatus, err
	}

	// Anything linked to this variable will transmit response headers.
	header := metadata.New(map[string]string{
		"username":  request.Username,
		"syb-token": token,
	})

	if err := grpc.SendHeader(ctx, header); err != nil {
		return falseStatus, err
	}

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) ChangePassword(ctx context.Context, request *ChangeCredentials) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateChangePassword(request.Username, request.OldPassword, request.NewPassword); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_CHANGE_PASSWORD)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.CHANGE_PASSWORD, structs.ChangeCredentials{
		Username:    request.Username,
		OldPassword: auth.GenerateHash(request.OldPassword, s.salt),
		NewPassword: auth.GenerateHash(request.NewPassword, s.salt),
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Nack(ctx context.Context, request *AckUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateAck(request.QueueName, request.Id); err != nil {
		return falseStatus, err
	}

	if _, err := uuid.FromBytes(request.Id); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	consumerID, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasPermission(username, request.QueueName, rbac.ACK_ACTION)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.NACK, structs.AckUpdate{
		QueueName:  request.QueueName,
		Id:         request.Id,
		ConsumerID: consumerID,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteQueue(ctx context.Context, request *DeleteQueueInfo) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if len(request.QueueName) == 0 {
		return falseStatus, ErrInvalidQueueName
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_DELETE_QUEUE)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.DELETE_QUEUE, structs.DeleteQueueInfo{
		QueueName: request.QueueName,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) AddRoutingKey(ctx context.Context, in *AddRoute) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateRoutingKeyChange(in.QueueName, in.RouteName); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_ADD_ROUTING_KEY)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.ADD_ROUTING_KEY, structs.AddRoute{
		RouteName: in.RouteName,
		QueueName: in.QueueName,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteRoutingKey(ctx context.Context, in *DeleteRoute) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateRoutingKeyChange(in.QueueName, in.RouteName); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_DELETE_ROUTING_KEY)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.DELETE_ROUTING_KEY, structs.DeleteRoute{
		RouteName: in.RouteName,
		QueueName: in.QueueName,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) CreateUser(ctx context.Context, in *UserCreds) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateUsernamePassword(in.Username, in.Password); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_CREATE_USER)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.CREATE_ACCOUNT, structs.UserCreds{
		Username: in.Username,
		Password: auth.GenerateHash(in.Password, s.salt),
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

type prebatch struct {
	data [][]byte
	id   [][]byte
}

func (s mQEndpointsServer) SubmitBatchedMessages(ctx context.Context, in *BatchMessages) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	routes := make(map[string]*prebatch)
	for _, msg := range in.Messages {
		if err := validateMessage(msg.Data, msg.RoutingKey); err != nil {
			return falseStatus, err
		}

		if _, ok := routes[msg.RoutingKey]; ok {
			id := uuid.New()
			idRaw := id[:]

			routes[msg.RoutingKey].data = append(routes[msg.RoutingKey].data, msg.Data)
			routes[msg.RoutingKey].id = append(routes[msg.RoutingKey].id, idRaw)

			continue
		}

		// only check it the first time we see the routing key
		ok, err = s.rbacManager.HasPermission(username, msg.RoutingKey, rbac.SUBMIT_BATCH_ACTION)
		if err != nil {
			return falseStatus, err
		}

		if !ok {
			return falseStatus, fmt.Errorf("does not have permission to perform action")
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

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	_, err = s.sendCommand(fsm.SUBMIT_BATCH_MESSAGE, structs.BatchMessages{
		Messages: msgs,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) IsLeaderNode(ctx context.Context, msg *LeaderNodeRequest) (*Status, error) {
	if s.raftServer == nil {
		return falseStatus, ErrRaftNotIntialised
	}

	return &Status{
		Status: s.raftServer.State() == raft.LEADER,
	}, nil
}

func (s mQEndpointsServer) DeleteUser(ctx context.Context, msg *UserInformation) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if len(msg.Username) == 0 {
		return falseStatus, ErrInvalidUsername
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_DELETE_USER)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.DELETE_USER, structs.UserInformation{
		Username: msg.Username,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchAck(ctx context.Context, msg *BatchAckUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateBatchAck(msg.QueueName, msg.Id); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	cnId, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasPermission(username, msg.QueueName, rbac.BATCH_ACK_ACTION)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.BATCH_ACK, structs.BatchAckUpdate{
		QueueName:  msg.QueueName,
		Ids:        msg.Id,
		ConsumerID: cnId,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchNack(ctx context.Context, msg *BatchNackUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return falseStatus, ErrNotLeader
	}

	if err := validateBatchAck(msg.QueueName, msg.Ids); err != nil {
		return falseStatus, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return falseStatus, ErrMetaDataNotFound
	}

	cnId, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return falseStatus, ErrNoAuthToken
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return falseStatus, err
	}

	ok, err = s.rbacManager.HasPermission(username, msg.QueueName, rbac.BATCH_ACK_ACTION)
	if err != nil {
		return falseStatus, err
	}

	if !ok {
		return falseStatus, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.BATCH_ACK, structs.BatchNackUpdate{
		QueueName:  msg.QueueName,
		Ids:        msg.Ids,
		ConsumerID: cnId,
	}, username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) LogOut(ctx context.Context, req *LogOutRequest) (*LogOutResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &LogOutResponse{
			Status: false,
		}, ErrMetaDataNotFound
	}

	err := s.authManager.LogOut(&md)

	return &LogOutResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) CreateRole(ctx context.Context, req *CreateRoleRequest) (*CreateRoleResponse, error) {
	if s.raftServer.State() != raft.LEADER {
		return &CreateRoleResponse{
			Status: false,
		}, ErrNotLeader
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &CreateRoleResponse{
			Status: false,
		}, ErrMetaDataNotFound
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return &CreateRoleResponse{
			Status: false,
		}, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_CREATE_ROLE)
	if err != nil {
		return &CreateRoleResponse{
			Status: false,
		}, err
	}

	if !ok {
		return &CreateRoleResponse{
			Status: false,
		}, fmt.Errorf("does not have permission to perform action")
	}

	roleConvert, err := s.rbacManager.CreateRole(req.Role)
	if err != nil {
		return &CreateRoleResponse{
			Status: false,
		}, err
	}

	_, err = s.sendCommand(fsm.CREATE_ROLE, roleConvert, username)

	return &CreateRoleResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) AssignRole(ctx context.Context, req *AssignRoleRequest) (*AssignRoleResponse, error) {
	if s.raftServer.State() != raft.LEADER {
		return &AssignRoleResponse{
			Status: false,
		}, ErrNotLeader
	}

	if req == nil || len(req.Role) == 0 || len(req.Username) == 0 {
		return &AssignRoleResponse{
			Status: false,
		}, fmt.Errorf("invalid request payload for AssignRole")
	}

	if !s.authManager.UserExists(req.Username) {
		return &AssignRoleResponse{
			Status: false,
		}, fmt.Errorf("user with name '%s' does not exist", req.Username)
	}

	if !s.rbacManager.RoleExists(req.Role) {
		return &AssignRoleResponse{
			Status: false,
		}, fmt.Errorf("role with name '%s' does not exist", req.Role)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &AssignRoleResponse{
			Status: false,
		}, ErrMetaDataNotFound
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return &AssignRoleResponse{
			Status: false,
		}, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_ASSIGN_ROLE)
	if err != nil {
		return &AssignRoleResponse{
			Status: false,
		}, err
	}

	if !ok {
		return &AssignRoleResponse{
			Status: false,
		}, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.ASSIGN_ROLE, structs.RoleUsername{
		Role:     req.Role,
		Username: req.Username,
	}, username)

	return &AssignRoleResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) UnassignRole(ctx context.Context, req *UnassignRoleRequest) (*UnassignRoleResponse, error) {
	if s.raftServer.State() != raft.LEADER {
		return &UnassignRoleResponse{
			Status: false,
		}, ErrNotLeader
	}

	if req == nil || len(req.Role) == 0 || len(req.Username) == 0 {
		return &UnassignRoleResponse{
			Status: false,
		}, fmt.Errorf("invalid request payload for AssignRole")
	}

	if !s.authManager.UserExists(req.Username) {
		return &UnassignRoleResponse{
			Status: false,
		}, fmt.Errorf("user with name '%s' does not exist", req.Username)
	}

	if !s.rbacManager.RoleExists(req.Role) {
		return &UnassignRoleResponse{
			Status: false,
		}, fmt.Errorf("role with name '%s' does not exist", req.Role)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &UnassignRoleResponse{
			Status: false,
		}, ErrMetaDataNotFound
	}

	username, err := s.authManager.GetUsername(&md)
	if err != nil {
		log.Err(err).Msg("unable to get username")
		return &UnassignRoleResponse{
			Status: false,
		}, err
	}

	ok, err = s.rbacManager.HasAdminPermission(username, rbac.ALLOW_ASSIGN_ROLE)
	if err != nil {
		return &UnassignRoleResponse{
			Status: false,
		}, err
	}

	if !ok {
		return &UnassignRoleResponse{
			Status: false,
		}, fmt.Errorf("does not have permission to perform action")
	}

	_, err = s.sendCommand(fsm.UNASSIGN_ROLE, structs.RoleUsername{
		Role:     req.Role,
		Username: req.Username,
	}, username)

	return &UnassignRoleResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) sendCommand(payloadType fsm.Operation, payload interface{}, username string) (interface{}, error) {
	jsonBytes, err := msgpack.Marshal(payload)
	if err != nil {
		return &fsm.ApplyResponse{}, err
	}

	obj := s.getCommandPool.GetObject()
	obj.Op = payloadType
	obj.Data = jsonBytes
	obj.Username = username

	data, err := msgpack.Marshal(obj)

	if err != nil {
		return &fsm.ApplyResponse{}, err
	}

	return s.raftServer.ApplyLog(&data, raft.DATA_LOG)
}
