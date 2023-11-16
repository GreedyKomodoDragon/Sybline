package auth

import (
	"context"
	"time"
)

type FollowerSessions interface {
	PropagateSessions(token string, username string, id []byte, ttl time.Time) error
	DeleteSessions(token string, username string) error
}

type followerSessions struct {
	clients []SessionClient
}

func NewFollowerSessions(clients []SessionClient) FollowerSessions {
	return &followerSessions{
		clients: clients,
	}
}

func (f *followerSessions) PropagateSessions(token string, username string, id []byte, ttl time.Time) error {
	payload := &PropagateSessionRequest{
		Token:      token,
		Username:   username,
		ConsumerId: id,
		TimeToLive: ttl.Unix(),
	}

	for _, client := range f.clients {
		if _, err := client.PropagateSession(context.Background(), payload); err != nil {
			return err
		}
	}
	return nil
}

func (f *followerSessions) DeleteSessions(token string, username string) error {
	payload := &DeleteSessionRequest{
		Token:    token,
		Username: username,
	}

	for _, client := range f.clients {
		if _, err := client.DeleteSession(context.Background(), payload); err != nil {
			return err
		}
	}

	return nil
}
