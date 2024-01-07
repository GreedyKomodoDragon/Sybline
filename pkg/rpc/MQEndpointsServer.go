package rpc

import (
	"context"
	"errors"
	"sybline/pkg/handler"

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
	hand handler.Handler
	UnimplementedMQEndpointsServer
}

func NewServer(hand handler.Handler) MQEndpointsServer {
	return mQEndpointsServer{
		hand: hand,
	}
}

func (s mQEndpointsServer) SubmitMessage(ctx context.Context, info *MessageInfo) (*Status, error) {
	err := s.hand.SubmitMessage(ctx, info.RoutingKey, info.Data)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) GetMessages(ctx context.Context, request *RequestMessageData) (*MessageCollection, error) {
	messages := []*MessageData{}

	msgs, err := s.hand.GetMessages(ctx, request.QueueName, request.Amount)
	if err != nil {
		return &MessageCollection{
			Messages: messages,
		}, err
	}

	// Convert between APIs
	for _, msg := range msgs {
		messages = append(messages, &MessageData{
			Id:   msg.Id,
			Data: msg.Data,
		})
	}

	return &MessageCollection{
		Messages: messages,
	}, nil
}

func (s mQEndpointsServer) CreateQueue(ctx context.Context, req *QueueInfo) (*Status, error) {
	err := s.hand.CreateQueue(ctx, req.RoutingKey, req.Name, req.Size, req.RetryLimit, req.HasDLQueue)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Ack(ctx context.Context, req *AckUpdate) (*Status, error) {
	err := s.hand.Ack(ctx, req.QueueName, req.Id)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Login(ctx context.Context, request *Credentials) (*Status, error) {
	token, err := s.hand.Login(ctx, request.Username, request.Password)
	if err != nil {
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
	err := s.hand.ChangePassword(ctx, request.Username, request.OldPassword, request.NewPassword)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) Nack(ctx context.Context, request *AckUpdate) (*Status, error) {
	err := s.hand.Nack(ctx, request.QueueName, request.Id)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteQueue(ctx context.Context, request *DeleteQueueInfo) (*Status, error) {
	err := s.hand.DeleteQueue(ctx, request.QueueName)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) AddRoutingKey(ctx context.Context, request *AddRoute) (*Status, error) {
	err := s.hand.AddRoutingKey(ctx, request.RouteName, request.QueueName)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) DeleteRoutingKey(ctx context.Context, request *DeleteRoute) (*Status, error) {
	err := s.hand.DeleteRoutingKey(ctx, request.RouteName, request.QueueName)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) CreateUser(ctx context.Context, request *UserCreds) (*Status, error) {
	err := s.hand.CreateUser(ctx, request.Username, request.Password)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) SubmitBatchedMessages(ctx context.Context, request *BatchMessages) (*Status, error) {
	messages := []handler.MessageSubmit{}

	for _, msg := range request.Messages {
		messages = append(messages, handler.MessageSubmit{
			RoutingKey: msg.RoutingKey,
			Data:       msg.Data,
		})
	}

	err := s.hand.SubmitBatchedMessages(ctx, messages)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) IsLeaderNode(ctx context.Context, msg *LeaderNodeRequest) (*Status, error) {
	ok, err := s.hand.IsLeaderNode(ctx)
	if err != nil {
		return falseStatus, err
	}

	return &Status{
		Status: ok,
	}, nil
}

func (s mQEndpointsServer) DeleteUser(ctx context.Context, request *UserInformation) (*Status, error) {
	err := s.hand.DeleteUser(ctx, request.Username)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchAck(ctx context.Context, request *BatchAckUpdate) (*Status, error) {
	err := s.hand.BatchAck(ctx, request.QueueName, request.Id)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) BatchNack(ctx context.Context, request *BatchNackUpdate) (*Status, error) {
	err := s.hand.BatchNack(ctx, request.QueueName, request.Ids)

	return &Status{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) LogOut(ctx context.Context, request *LogOutRequest) (*LogOutResponse, error) {
	err := s.hand.LogOut(ctx)

	return &LogOutResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) CreateRole(ctx context.Context, request *CreateRoleRequest) (*CreateRoleResponse, error) {
	err := s.hand.CreateRole(ctx, request.Role)

	return &CreateRoleResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) AssignRole(ctx context.Context, request *AssignRoleRequest) (*AssignRoleResponse, error) {
	err := s.hand.AssignRole(ctx, request.Role, request.Username)

	return &AssignRoleResponse{
		Status: err == nil,
	}, err
}

func (s mQEndpointsServer) UnassignRole(ctx context.Context, request *UnassignRoleRequest) (*UnassignRoleResponse, error) {
	err := s.hand.UnassignRole(ctx, request.Role, request.Username)

	return &UnassignRoleResponse{
		Status: err == nil,
	}, err
}
