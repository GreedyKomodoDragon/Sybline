package handler

import (
	"context"
	"errors"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/fsm"
	"sybline/pkg/messages"
	"sybline/pkg/structs"

	"time"

	raft "github.com/GreedyKomodoDragon/raft"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc"
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

type MQEndpointsServer interface {
	SubmitMessage(context.Context, *messages.MessageInfo) (*messages.Status, error)
	GetMessages(context.Context, *messages.RequestMessageData) (*messages.MessageCollection, error)
	CreateQueue(context.Context, *messages.QueueInfo) (*messages.Status, error)
	Ack(context.Context, *messages.AckUpdate) (*messages.Status, error)
	Login(context.Context, *messages.Credentials) (*messages.Status, error)
	ChangePassword(context.Context, *messages.ChangeCredentials) (*messages.Status, error)
	Nack(context.Context, *messages.AckUpdate) (*messages.Status, error)
	DeleteQueue(context.Context, *messages.DeleteQueueInfo) (*messages.Status, error)
	AddRoutingKey(context.Context, *messages.AddRoute) (*messages.Status, error)
	DeleteRoutingKey(context.Context, *messages.DeleteRoute) (*messages.Status, error)
	CreateUser(context.Context, *messages.UserCreds) (*messages.Status, error)
	SubmitBatchedMessages(context.Context, *messages.BatchMessages) (*messages.Status, error)
	IsLeaderNode(context.Context, *messages.LeaderNodeRequest) (*messages.Status, error)
	DeleteUser(context.Context, *messages.UserInformation) (*messages.Status, error)
	BatchAck(context.Context, *messages.BatchAckUpdate) (*messages.Status, error)
	BatchNack(context.Context, *messages.BatchNackUpdate) (*messages.Status, error)

	mustEmbedUnimplementedMQEndpointsServer()
}
type mQEndpointsServer struct {
	authManager    auth.AuthManager
	raftServer     raft.Raft
	salt           string
	getObjectPool  *ObjectPool[structs.RequestMessageData]
	getCommandPool *ObjectPool[fsm.CommandPayload]
	unimplementedMQEndpointsServer
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

func (s mQEndpointsServer) SubmitMessage(ctx context.Context, info *messages.MessageInfo) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) GetMessages(ctx context.Context, request *messages.RequestMessageData) (*messages.MessageCollection, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.MessageCollection{}, ErrNotLeader
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
	msgConverted := make([]*messages.MessageData, len(msgs))

	for index, msg := range msgs {
		msgConverted[index] = &messages.MessageData{
			Id:   msg.Id,
			Data: msg.Data,
		}
	}

	return &messages.MessageCollection{
		Messages: msgConverted,
	}, nil
}

func (s mQEndpointsServer) CreateQueue(ctx context.Context, request *messages.QueueInfo) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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
		return nil, ErrNoAuthToken
	}

	_, err := s.sendCommand(fsm.CREATE_QUEUE, structs.QueueInfo{
		RoutingKey: request.RoutingKey,
		Name:       request.Name,
		Size:       request.Size,
		RetryLimit: request.RetryLimit,
		HasDLQueue: request.HasDLQueue,
	})

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Ack(ctx context.Context, request *messages.AckUpdate) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Login(ctx context.Context, request *messages.Credentials) (*messages.Status, error) {
	if len(request.Username) == 0 {
		return nil, ErrInvalidUsername
	}

	if len(request.Password) == 0 {
		return nil, ErrInvalidPassword
	}

	token, err := s.authManager.Login(request.Username, auth.GenerateHash(request.Password, s.salt))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Anything linked to this variable will transmit response headers.
	header := metadata.New(map[string]string{
		"username":  request.Username,
		"syb-token": token,
	})

	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, err
	}

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) ChangePassword(ctx context.Context, request *messages.ChangeCredentials) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Nack(ctx context.Context, request *messages.AckUpdate) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteQueue(ctx context.Context, request *messages.DeleteQueueInfo) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) AddRoutingKey(ctx context.Context, in *messages.AddRoute) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteRoutingKey(ctx context.Context, in *messages.DeleteRoute) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) CreateUser(ctx context.Context, in *messages.UserCreds) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

type prebatch struct {
	data [][]byte
	id   [][]byte
}

func (s mQEndpointsServer) SubmitBatchedMessages(ctx context.Context, in *messages.BatchMessages) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) IsLeaderNode(ctx context.Context, msg *messages.LeaderNodeRequest) (*messages.Status, error) {
	if s.raftServer == nil {
		return nil, ErrRaftNotIntialised
	}

	return &messages.Status{
		Status: s.raftServer.State() == raft.LEADER,
	}, nil
}

func (s mQEndpointsServer) DeleteUser(ctx context.Context, msg *messages.UserInformation) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchAck(ctx context.Context, msg *messages.BatchAckUpdate) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchNack(ctx context.Context, msg *messages.BatchNackUpdate) (*messages.Status, error) {
	if s.raftServer.State() != raft.LEADER {
		return &messages.Status{
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

	return &messages.Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) sendCommand(payloadType fsm.Operation, payload interface{}) (interface{}, error) {
	obj := s.getCommandPool.GetObject()
	obj.Data = payload
	obj.Op = payloadType

	data, err := msgpack.Marshal(obj)
	if err != nil {
		return &fsm.ApplyResponse{}, err
	}

	return s.raftServer.ApplyLog(&data, raft.DATA_LOG)
}

// UnsafeMQEndpointsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MQEndpointsServer will
// result in compilation errors.
type UnsafeMQEndpointsServer interface {
	mustEmbedUnimplementedMQEndpointsServer()
}

func RegisterMQEndpointsServer(s grpc.ServiceRegistrar, srv MQEndpointsServer) {
	s.RegisterService(&MQEndpoints_ServiceDesc, srv)
}

func _MQEndpoints_SubmitMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.MessageInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).SubmitMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/SubmitMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).SubmitMessage(ctx, req.(*messages.MessageInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_GetMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.RequestMessageData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).GetMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/GetMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).GetMessages(ctx, req.(*messages.RequestMessageData))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_CreateQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.QueueInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).CreateQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/CreateQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).CreateQueue(ctx, req.(*messages.QueueInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_Ack_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.AckUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).Ack(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/Ack",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).Ack(ctx, req.(*messages.AckUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.Credentials)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).Login(ctx, req.(*messages.Credentials))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.ChangeCredentials)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).ChangePassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/ChangePassword",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).ChangePassword(ctx, req.(*messages.ChangeCredentials))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_Nack_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.AckUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).Nack(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/Nack",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).Nack(ctx, req.(*messages.AckUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_DeleteQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.DeleteQueueInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).DeleteQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/DeleteQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).DeleteQueue(ctx, req.(*messages.DeleteQueueInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_AddRoutingKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.AddRoute)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).AddRoutingKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/AddRoutingKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).AddRoutingKey(ctx, req.(*messages.AddRoute))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_DeleteRoutingKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.DeleteRoute)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).DeleteRoutingKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/DeleteRoutingKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).DeleteRoutingKey(ctx, req.(*messages.DeleteRoute))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_CreateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.UserCreds)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).CreateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/CreateUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).CreateUser(ctx, req.(*messages.UserCreds))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_SubmitBatchedMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.BatchMessages)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).SubmitBatchedMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/SubmitBatchedMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).SubmitBatchedMessages(ctx, req.(*messages.BatchMessages))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_IsLeaderNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.LeaderNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).IsLeaderNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/IsLeaderNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).IsLeaderNode(ctx, req.(*messages.LeaderNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.UserInformation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/DeleteUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).DeleteUser(ctx, req.(*messages.UserInformation))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_BatchAck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.BatchAckUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).BatchAck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/BatchAck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).BatchAck(ctx, req.(*messages.BatchAckUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _MQEndpoints_BatchNack_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(messages.BatchNackUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MQEndpointsServer).BatchNack(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/MQ.MQEndpoints/BatchNack",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MQEndpointsServer).BatchNack(ctx, req.(*messages.BatchNackUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

// MQEndpoints_ServiceDesc is the grpc.ServiceDesc for MQEndpoints service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MQEndpoints_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "MQ.MQEndpoints",
	HandlerType: (*MQEndpointsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitMessage",
			Handler:    _MQEndpoints_SubmitMessage_Handler,
		},
		{
			MethodName: "GetMessages",
			Handler:    _MQEndpoints_GetMessages_Handler,
		},
		{
			MethodName: "CreateQueue",
			Handler:    _MQEndpoints_CreateQueue_Handler,
		},
		{
			MethodName: "Ack",
			Handler:    _MQEndpoints_Ack_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _MQEndpoints_Login_Handler,
		},
		{
			MethodName: "ChangePassword",
			Handler:    _MQEndpoints_ChangePassword_Handler,
		},
		{
			MethodName: "Nack",
			Handler:    _MQEndpoints_Nack_Handler,
		},
		{
			MethodName: "DeleteQueue",
			Handler:    _MQEndpoints_DeleteQueue_Handler,
		},
		{
			MethodName: "AddRoutingKey",
			Handler:    _MQEndpoints_AddRoutingKey_Handler,
		},
		{
			MethodName: "DeleteRoutingKey",
			Handler:    _MQEndpoints_DeleteRoutingKey_Handler,
		},
		{
			MethodName: "CreateUser",
			Handler:    _MQEndpoints_CreateUser_Handler,
		},
		{
			MethodName: "SubmitBatchedMessages",
			Handler:    _MQEndpoints_SubmitBatchedMessages_Handler,
		},
		{
			MethodName: "IsLeaderNode",
			Handler:    _MQEndpoints_IsLeaderNode_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _MQEndpoints_DeleteUser_Handler,
		},
		{
			MethodName: "BatchAck",
			Handler:    _MQEndpoints_BatchAck_Handler,
		},
		{
			MethodName: "BatchNack",
			Handler:    _MQEndpoints_BatchNack_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mq.proto",
}
