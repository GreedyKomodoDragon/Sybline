package auth

import (
	"errors"
	"sync"
	"time"
)

var ErrNoUserSessionExists = errors.New("unable to find user")
var ErrTokenDoesNotExist = errors.New("token does not exist")

type SessionHandler interface {
	GetConsumerID(token string, username string) ([]byte, error)
	RemoveToken(token string, username string) error
	AddToken(token string, username string, consumerID []byte, ttl time.Time) error
	DeleteUser(username string) error
}

type timeSession struct {
	time  time.Time
	id    []byte
	token string
}

func NewSessionHandler() SessionHandler {
	return &sessionHandler{
		lock:       &sync.RWMutex{},
		sessionIDs: map[string][]timeSession{},
	}
}

type sessionHandler struct {
	lock       *sync.RWMutex
	sessionIDs map[string][]timeSession
}

func remove(s []timeSession, i int) []timeSession {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (r *sessionHandler) GetConsumerID(token string, username string) ([]byte, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// use local session if possible
	sess, ok := r.sessionIDs[username]
	if !ok {
		return nil, ErrNoUserSessionExists
	}

	for _, value := range sess {
		if value.token == token {
			if t := time.Now(); value.time.After(t) {
				id := value.id
				return id, nil
			}

			// makes locking easier to manage if concurrent request
			go r.RemoveToken(username, token)

			return nil, ErrNoUserSessionExists
		}
	}

	return nil, ErrNoUserSessionExists
}

func (r *sessionHandler) RemoveToken(token string, username string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	// use local session if possible
	sess, ok := r.sessionIDs[username]
	if !ok {
		return ErrUserDoesNotExist
	}

	for index, value := range sess {
		if value.token == token {

			if len(sess) == 1 {
				delete(r.sessionIDs, username)
			} else {
				sess = remove(sess, index)
			}

			return nil
		}
	}

	return ErrTokenDoesNotExist
}

func (r *sessionHandler) AddToken(token string, username string, consumerID []byte, ttl time.Time) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if mp, ok := r.sessionIDs[username]; ok {
		mp = append(mp, timeSession{
			time:  ttl,
			id:    consumerID,
			token: token,
		})

		r.sessionIDs[username] = mp
		return nil
	}

	r.sessionIDs[username] = []timeSession{
		{
			time:  ttl,
			id:    consumerID,
			token: token,
		},
	}

	return nil
}

func (r *sessionHandler) DeleteUser(username string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.sessionIDs, username)

	return nil
}
