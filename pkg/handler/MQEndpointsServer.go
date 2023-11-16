package handler

import (
	"context"
	"errors"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/fsm"
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
)

type mQEndpointsServer struct {
	authManager    auth.AuthManager
	raftServer     raft.Raft
	salt           string
	getObjectPool  *ObjectPool[structs.RequestMessageData]
	getCommandPool *ObjectPool[fsm.CommandPayload]
	UnimplementedMQEndpointsServer
}

func NewServer(authManager auth.AuthManager, raftServer raft.Raft, salt string) MQEndpointsServer {
	return mQEndpointsServer{
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
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateMessage(info.Data, info.RoutingKey); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	_, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return nil, ErrNoAuthToken
	}

	id := uuid.New()

	_, err = s.sendCommand(fsm.SUBMIT_MESSAGE, structs.MessageInfo{
		Rk:   info.RoutingKey,
		Data: info.Data,
		Id:   id[:],
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) GetMessages(ctx context.Context, request *RequestMessageData) (*MessageCollection, error) {
	if s.raftServer.State() != raft.LEADER {
		return &MessageCollection{}, ErrNotLeader
	}

	if err := validateRequestMessages(request.QueueName, request.Amount); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	consumerID, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return nil, ErrNoAuthToken
	}

	msg := s.getObjectPool.GetObject()
	msg.QueueName = request.QueueName
	msg.Amount = request.Amount
	msg.ConsumerID = consumerID
	msg.Time = time.Now().Unix()

	res, err := s.sendCommand(fsm.GET_MESSAGES, msg)

	s.getObjectPool.ReleaseObject(msg)
	if err != nil {
		return nil, err
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
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateCreateQueue(request.RoutingKey, request.Name, request.Size); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		log.Err(err).Msg("unable to get consumer id")
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.CREATE_QUEUE, structs.QueueInfo{
		RoutingKey: request.RoutingKey,
		Name:       request.Name,
		Size:       request.Size,
		RetryLimit: request.RetryLimit,
		HasDLQueue: request.HasDLQueue,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Ack(ctx context.Context, request *AckUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateAck(request.QueueName, request.Id); err != nil {
		return nil, err
	}

	if _, err := uuid.FromBytes(request.Id); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	consumerID, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return nil, ErrNoAuthToken
	}

	_, err = s.sendCommand(fsm.ACK, structs.AckUpdate{
		QueueName:  request.QueueName,
		Id:         request.Id,
		ConsumerID: consumerID,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Login(ctx context.Context, request *Credentials) (*Status, error) {
	if len(request.Username) == 0 {
		return nil, ErrInvalidUsername
	}

	if len(request.Password) == 0 {
		return nil, ErrInvalidPassword
	}

	token, err := s.authManager.Login(request.Username, auth.GenerateHash(request.Password, s.salt))
	if err != nil {
		log.Err(err).Msg("unable to log in")
		return nil, err
	}

	// Anything linked to this variable will transmit response headers.
	header := metadata.New(map[string]string{
		"username":  request.Username,
		"syb-token": token,
	})

	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, err
	}

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) ChangePassword(ctx context.Context, request *ChangeCredentials) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateChangePassword(request.Username, request.OldPassword, request.NewPassword); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.CHANGE_PASSWORD, structs.ChangeCredentials{
		Username:    request.Username,
		OldPassword: auth.GenerateHash(request.OldPassword, s.salt),
		NewPassword: auth.GenerateHash(request.NewPassword, s.salt),
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Nack(ctx context.Context, request *AckUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateAck(request.QueueName, request.Id); err != nil {
		return nil, err
	}

	if _, err := uuid.FromBytes(request.Id); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	consumerID, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return nil, ErrNoAuthToken
	}

	_, err = s.sendCommand(fsm.NACK, structs.AckUpdate{
		QueueName:  request.QueueName,
		Id:         request.Id,
		ConsumerID: consumerID,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteQueue(ctx context.Context, request *DeleteQueueInfo) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if len(request.QueueName) == 0 {
		return nil, ErrInvalidQueueName
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.DELETE_QUEUE, structs.DeleteQueueInfo{
		QueueName: request.QueueName,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) AddRoutingKey(ctx context.Context, in *AddRoute) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateRoutingKeyChange(in.QueueName, in.RouteName); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.ADD_ROUTING_KEY, structs.AddRoute{
		RouteName: in.RouteName,
		QueueName: in.QueueName,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteRoutingKey(ctx context.Context, in *DeleteRoute) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateRoutingKeyChange(in.QueueName, in.RouteName); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.DELETE_ROUTING_KEY, structs.DeleteRoute{
		RouteName: in.RouteName,
		QueueName: in.QueueName,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) CreateUser(ctx context.Context, in *UserCreds) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateUsernamePassword(in.Username, in.Password); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.CREATE_ACCOUNT, structs.UserCreds{
		Username: in.Username,
		Password: auth.GenerateHash(in.Password, s.salt),
	})

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
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	routes := make(map[string]*prebatch)
	for _, msg := range in.Messages {
		if err := validateMessage(msg.Data, msg.RoutingKey); err != nil {
			return nil, err
		}

		id := uuid.New()
		idRaw := id[:]

		if _, ok := routes[msg.RoutingKey]; ok {
			routes[msg.RoutingKey].data = append(routes[msg.RoutingKey].data, msg.Data)
			routes[msg.RoutingKey].id = append(routes[msg.RoutingKey].id, idRaw)

			continue
		}

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
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.SUBMIT_BATCH_MESSAGE, structs.BatchMessages{
		Messages: msgs,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) IsLeaderNode(ctx context.Context, msg *LeaderNodeRequest) (*Status, error) {
	if s.raftServer == nil {
		return nil, ErrRaftNotIntialised
	}

	return &Status{
		Status: s.raftServer.State() == raft.LEADER,
	}, nil
}

func (s mQEndpointsServer) DeleteUser(ctx context.Context, msg *UserInformation) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if len(msg.Username) == 0 {
		return nil, ErrInvalidUsername
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	if _, err := s.authManager.GetConsumerID(&md); err != nil {
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.DELETE_USER, structs.UserInformation{
		Username: msg.Username,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchAck(ctx context.Context, msg *BatchAckUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateBatchAck(msg.QueueName, msg.Id); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	cnId, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return nil, ErrNoAuthToken
	}

	_, err = s.sendCommand(fsm.BATCH_ACK, structs.BatchAckUpdate{
		QueueName:  msg.QueueName,
		Ids:        msg.Id,
		ConsumerID: cnId,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchNack(ctx context.Context, msg *BatchNackUpdate) (*Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &Status{
			Status: false,
		}, ErrNotLeader
	}

	if err := validateBatchAck(msg.QueueName, msg.Ids); err != nil {
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	cnId, err := s.authManager.GetConsumerID(&md)
	if err != nil {
		return nil, ErrNoAuthToken
	}

	_, err = s.sendCommand(fsm.BATCH_ACK, structs.BatchNackUpdate{
		QueueName:  msg.QueueName,
		Ids:        msg.Ids,
		ConsumerID: cnId,
	})

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) LogOut(ctx context.Context, req *LogOutRequest) (*LogOutResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetaDataNotFound
	}

	err := s.authManager.LogOut(&md)

	return &LogOutResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) sendCommand(payloadType fsm.Operation, payload interface{}) (interface{}, error) {
	jsonBytes, err := msgpack.Marshal(payload)
	if err != nil {
		return &fsm.ApplyResponse{}, err
	}

	obj := s.getCommandPool.GetObject()
	obj.Op = payloadType

	data, err := msgpack.Marshal(fsm.CommandPayload{
		Op:   payloadType,
		Data: jsonBytes,
	})

	obj.Data = jsonBytes

	if err != nil {
		return &fsm.ApplyResponse{}, err
	}

	return s.raftServer.ApplyLog(&data, raft.DATA_LOG)
}
