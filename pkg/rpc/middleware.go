package rpc

import (
	"context"
	"strings"
	"sybline/pkg/auth"

	raft "github.com/GreedyKomodoDragon/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func AuthenticationInterceptor(authManager auth.AuthManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !strings.HasPrefix(info.FullMethod, "/MQ.") {
			return handler(ctx, req)
		}

		// skip if no authentication is needed
		if info.FullMethod == "/MQ.MQEndpoints/Login" ||
			info.FullMethod == "/MQ.MQEndpoints/IsLeaderNode" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrMetaDataNotFound
		}

		id, err := authManager.GetConsumerIDViaMD(&md)
		if err != nil {
			return nil, ErrNoAuthToken
		}

		username, token, err := authManager.GetUsernameToken(&md)
		if err != nil {
			return nil, err
		}

		// hold onto this information for later
		ctx = context.WithValue(ctx, "consumerID", id)
		ctx = context.WithValue(ctx, "username", username)
		ctx = context.WithValue(ctx, "token", token)

		// Proceed with handling the request
		return handler(ctx, req)
	}
}

func IsLeader(raftServer raft.Raft) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !strings.HasPrefix(info.FullMethod, "/MQ.") {
			return handler(ctx, req)
		}

		// skip if no authentication is needed
		if info.FullMethod == "/MQ.MQEndpoints/IsLeaderNode" {
			return handler(ctx, req)
		}

		if raftServer.State() != raft.LEADER {
			return nil, ErrNotLeader
		}

		// Proceed with handling the request
		return handler(ctx, req)
	}
}
