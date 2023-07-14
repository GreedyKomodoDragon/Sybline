package handler

import (
	"context"
	"sybline/pkg/messages"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// unimplementedMQEndpointsServer must be embedded to have forward compatible implementations.
type unimplementedMQEndpointsServer struct {
}

func (unimplementedMQEndpointsServer) SubmitMessage(context.Context, *messages.MessageInfo) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitMessage not implemented")
}

func (unimplementedMQEndpointsServer) GetMessages(context.Context, *messages.RequestMessageData) (*messages.MessageCollection, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessages not implemented")
}

func (unimplementedMQEndpointsServer) CreateQueue(context.Context, *messages.QueueInfo) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateQueue not implemented")
}

func (unimplementedMQEndpointsServer) Ack(context.Context, *messages.AckUpdate) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ack not implemented")
}

func (unimplementedMQEndpointsServer) Login(context.Context, *messages.Credentials) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (unimplementedMQEndpointsServer) ChangePassword(context.Context, *messages.ChangeCredentials) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (unimplementedMQEndpointsServer) Delete(context.Context, *messages.DeleteQueueInfo) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method delete not implemented")
}

func (unimplementedMQEndpointsServer) DeleteRoutingKey(context.Context, *messages.DeleteRoute) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRoutingKey not implemented")
}

func (unimplementedMQEndpointsServer) CreateUser(context.Context, *messages.UserCreds) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}

func (unimplementedMQEndpointsServer) SubmitBatchedMessages(context.Context, *messages.BatchMessages) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitBatchedMessages not implemented")
}

func (unimplementedMQEndpointsServer) IsLeaderNode(context.Context, *messages.LeaderNodeRequest) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsLeaderNode not implemented")
}

func (unimplementedMQEndpointsServer) DeleteUser(context.Context, *messages.UserInformation) (*messages.Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}

func (unimplementedMQEndpointsServer) mustEmbedUnimplementedMQEndpointsServer() {}
