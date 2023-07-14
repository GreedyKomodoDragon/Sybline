package handler

import (
	"context"
	"sybline/pkg/messages"

	"google.golang.org/grpc"
)

type MQEndpointsClient interface {
	SubmitMessage(ctx context.Context, in *messages.MessageInfo, opts ...grpc.CallOption) (*messages.Status, error)
	GetMessages(ctx context.Context, in *messages.RequestMessageData, opts ...grpc.CallOption) (*messages.MessageCollection, error)
	CreateQueue(ctx context.Context, in *messages.QueueInfo, opts ...grpc.CallOption) (*messages.Status, error)
	Ack(ctx context.Context, in *messages.AckUpdate, opts ...grpc.CallOption) (*messages.Status, error)
	Login(ctx context.Context, in *messages.Credentials, opts ...grpc.CallOption) (*messages.Status, error)
	ChangePassword(ctx context.Context, in *messages.ChangeCredentials, opts ...grpc.CallOption) (*messages.Status, error)
	Nack(ctx context.Context, in *messages.AckUpdate, opts ...grpc.CallOption) (*messages.Status, error)
	DeleteOueue(ctx context.Context, in *messages.DeleteQueueInfo, opts ...grpc.CallOption) (*messages.Status, error)
	AddRoutingKey(ctx context.Context, in *messages.AddRoute, opts ...grpc.CallOption) (*messages.Status, error)
	DeleteRoutingKey(ctx context.Context, in *messages.DeleteRoute, opts ...grpc.CallOption) (*messages.Status, error)
	CreateUser(ctx context.Context, in *messages.UserCreds, opts ...grpc.CallOption) (*messages.Status, error)
	SubmitBatchedMessages(ctx context.Context, in *messages.BatchMessages, opts ...grpc.CallOption) (*messages.Status, error)
	IsLeaderNode(ctx context.Context, in *messages.LeaderNodeRequest, opts ...grpc.CallOption) (*messages.Status, error)
	DeleteUser(ctx context.Context, in *messages.UserInformation, opts ...grpc.CallOption) (*messages.Status, error)
	BatchAck(ctx context.Context, in *messages.BatchAckUpdate, opts ...grpc.CallOption) (*messages.Status, error)
	BatchNack(ctx context.Context, in *messages.BatchNackUpdate, opts ...grpc.CallOption) (*messages.Status, error)
}

type mQEndpointsClient struct {
	cc grpc.ClientConnInterface
}

func NewMQEndpointsClient(cc grpc.ClientConnInterface) MQEndpointsClient {
	return &mQEndpointsClient{cc}
}

func (c *mQEndpointsClient) SubmitMessage(ctx context.Context, in *messages.MessageInfo, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/SubmitMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) GetMessages(ctx context.Context, in *messages.RequestMessageData, opts ...grpc.CallOption) (*messages.MessageCollection, error) {
	out := new(messages.MessageCollection)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/GetMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) CreateQueue(ctx context.Context, in *messages.QueueInfo, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)

	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/CreateQueue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) Ack(ctx context.Context, in *messages.AckUpdate, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/Ack", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) Login(ctx context.Context, in *messages.Credentials, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) ChangePassword(ctx context.Context, in *messages.ChangeCredentials, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/ChangePassword", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) Nack(ctx context.Context, in *messages.AckUpdate, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/Nack", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) DeleteOueue(ctx context.Context, in *messages.DeleteQueueInfo, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/DeleteQueue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) AddRoutingKey(ctx context.Context, in *messages.AddRoute, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/AddRoutingKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) DeleteRoutingKey(ctx context.Context, in *messages.DeleteRoute, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/DeleteRoutingKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) CreateUser(ctx context.Context, in *messages.UserCreds, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/CreateUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) SubmitBatchedMessages(ctx context.Context, in *messages.BatchMessages, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/SubmitBatchedMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) IsLeaderNode(ctx context.Context, in *messages.LeaderNodeRequest, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/IsLeaderNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) DeleteUser(ctx context.Context, in *messages.UserInformation, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/DeleteUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) BatchAck(ctx context.Context, in *messages.BatchAckUpdate, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/BatchAck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mQEndpointsClient) BatchNack(ctx context.Context, in *messages.BatchNackUpdate, opts ...grpc.CallOption) (*messages.Status, error) {
	out := new(messages.Status)
	err := c.cc.Invoke(ctx, "/MQ.MQEndpoints/BatchNack", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
