package auth

import (
	"context"
	"time"
)

type sessionServer struct {
	sessionHandler SessionHandler
	UnimplementedSessionServer
}

func NewSessionServer(sh SessionHandler) SessionServer {
	return &sessionServer{
		sessionHandler: sh,
	}
}

func (s *sessionServer) PropagateSession(ctx context.Context, req *PropagateSessionRequest) (*PropagateSessionResponse, error) {
	err := s.sessionHandler.AddToken(req.Token, req.Username, req.ConsumerId, time.Unix(req.TimeToLive, 0))
	return &PropagateSessionResponse{
		Successful: err == nil,
	}, err

}

func (s *sessionServer) DeleteSession(ctx context.Context, req *DeleteSessionRequest) (*DeleteSessionResponse, error) {
	err := s.sessionHandler.RemoveToken(req.Token, req.Username)
	return &DeleteSessionResponse{
		Successful: err == nil,
	}, err
}
