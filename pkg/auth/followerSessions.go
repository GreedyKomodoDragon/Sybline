package auth

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
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

	for index, client := range f.clients {
		if _, err := client.PropagateSession(context.Background(), payload); err != nil {
			log.Err(err).Uint64("index", uint64(index)).Msg("failed to propagate the session to the follower")
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
