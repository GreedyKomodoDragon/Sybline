package auth_test

import (
	"github.com/stretchr/testify/mock"
)

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
