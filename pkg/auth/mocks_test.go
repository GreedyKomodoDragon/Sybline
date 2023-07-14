package auth_test

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type SessionMock struct {
	mock.Mock
}

func (o *SessionMock) GetConsumerID(token string, username string) ([]byte, error) {
	args := o.Called(token, username)
	return args.Get(0).([]byte), args.Error(1)
}

func (o *SessionMock) RemoveToken(token string, username string) error {
	args := o.Called(token, username)
	return args.Error(0)
}

func (o *SessionMock) AddToken(token string, username string, consumerID []byte, ttl time.Duration) error {
	args := o.Called(token, username, consumerID, ttl)
	return args.Error(0)
}

func (o *SessionMock) IsRefreshRequired(token, username string) (bool, error) {
	args := o.Called(token, username)
	return args.Bool(0), args.Error(1)
}

func (o *SessionMock) DeleteTokenContaining(token string) error {
	args := o.Called(token)
	return args.Error(0)
}

type TokenGenMock struct {
	mock.Mock
}

func (o *TokenGenMock) Generate() string {
	args := o.Called()
	return args.String(0)
}

type IdGenMock struct {
	mock.Mock
}

func (o *IdGenMock) Generate() []byte {
	args := o.Called()
	return args.Get(0).([]byte)
}
