package auth_test

import (
	"math/rand"
	"sybline/pkg/auth"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestAuthManager_CreateUser_Sync(t *testing.T) {
	tGenMock := &TokenGenMock{}
	idGenMock := &IdGenMock{}
	mockSession := &SessionMock{}

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Second*60)

	err := a.CreateUser("username", "password")
	assert.Nil(t, err, "should be able to create a new use")

	err = a.CreateUser("username", "password")
	assert.EqualError(t, err, auth.ErrUsernameAlreadyTaken.Error())
}

func TestAuthManager_CreateUser_Async(t *testing.T) {
	tGenMock := &TokenGenMock{}
	idGenMock := &IdGenMock{}
	mockSession := &SessionMock{}

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Second*60)

	var wg sync.WaitGroup
	var n atomic.Uint32

	createUser := func(name string, password string) {
		r := rand.Intn(50)
		time.Sleep(time.Duration(r) * time.Microsecond)

		if err := a.CreateUser(name, password); err != nil {
			n.Add(1)
		}
		wg.Done()
	}

	wg.Add(3)
	go createUser("username", "password")
	go createUser("username", "password")
	go createUser("username", "password")

	wg.Wait()

	assert.Equal(t, uint32(2), n.Load(), "should have two errors")
}

func TestAuthManager_Login(t *testing.T) {
	tGenMock := &TokenGenMock{}
	tGenMock.On("Generate").Return("token")

	idGenMock := &IdGenMock{}
	idGenMock.On("Generate").Return([]byte("ab"))

	mockSession := &SessionMock{}
	mockSession.On("AddToken", "token", "username", []byte("ab"), time.Minute).Return(nil)

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Second*60)

	hashedPassword := auth.GenerateHash("password", "salt")

	err := a.CreateUser("username", hashedPassword)
	assert.Nil(t, err, "should be able to create a new use")

	_, err = a.Login("username", hashedPassword)
	assert.Nil(t, err, "should be able to login with correct details")

	token, err := a.Login("username", "password1")
	assert.EqualError(t, err, auth.ErrInvalidLogin.Error())
	assert.Len(t, token, 0)
}

func TestAuthManager_IsLogin(t *testing.T) {
	tGenMock := &TokenGenMock{}
	tGenMock.On("Generate").Return("token")

	idGenMock := &IdGenMock{}
	idGenMock.On("Generate").Return([]byte("ab"))

	mockSession := &SessionMock{}
	mockSession.On("AddToken", "token", "username", []byte("ab"), time.Minute).Return(nil)
	mockSession.On("GetConsumerID", "token", "username").Return([]byte("a"), nil)

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Second*60)

	hashedPassword := auth.GenerateHash("password", "salt")

	err := a.CreateUser("username", hashedPassword)
	assert.Nil(t, err, "should be able to create a new use")

	token, err := a.Login("username", hashedPassword)
	assert.Nil(t, err, "should be able to login with correct details")

	md := metadata.New(map[string]string{"syb-token": token, "username": "username"})
	_, err = a.GetConsumerID(&md)
	assert.Nil(t, err)

	md = metadata.New(map[string]string{"syb-token": "token"})
	_, err = a.GetConsumerID(&md)
	assert.NotNil(t, err)
}

func TestAuthManager_Change_Password(t *testing.T) {
	tGenMock := &TokenGenMock{}
	tGenMock.On("Generate").Return("token")

	idGenMock := &IdGenMock{}
	idGenMock.On("Generate").Return([]byte("ab"))

	mockSession := &SessionMock{}
	mockSession.On("AddToken", "token", "username", []byte("ab"), time.Minute).Return(nil)

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Minute)

	err := a.CreateUser("username", "password")
	assert.Nil(t, err, "should be able to create a new use")

	_, err = a.Login("username", "password")
	assert.Nil(t, err, "should be able to login with correct details")

	status, err := a.ChangePassword("username", "password", "password1")
	assert.Nil(t, err, "should be able to change a password")
	assert.True(t, status, "should return true on changing password")

	token, err := a.Login("username", "password")
	assert.EqualError(t, err, auth.ErrInvalidLogin.Error())
	assert.Equal(t, 0, len(token))

	_, err = a.Login("username", "password1")
	assert.Nil(t, err, "should be able to login with new correct details")
}

func TestAuthManager_DeleteUser_User_Exists(t *testing.T) {
	tGenMock := &TokenGenMock{}
	idGenMock := &IdGenMock{}
	mockSession := &SessionMock{}
	mockSession.On("DeleteTokenContaining", "username").Return(nil)
	mockSession.On("DeleteTokenContaining", "usernameOne").Return(nil)

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Second*60)

	err := a.CreateUser("username", "hashedPassword")
	assert.Nil(t, err, "should be able to create a new use")

	err = a.CreateUser("usernameOne", "hashedPassword")
	assert.Nil(t, err, "should be able to create a new use")

	err = a.DeleteUser("usernameOne")
	assert.Nil(t, err, "delete user should return no error")

	token, err := a.Login("usernameOne", "hashedPassword")
	assert.Equal(t, 0, len(token))
	assert.NotNil(t, err, "should not be able to log into deleted user")
}

func TestAuthManager_DeleteUser_User_Does_Not_Exist(t *testing.T) {
	tGenMock := &TokenGenMock{}
	idGenMock := &IdGenMock{}
	mockSession := &SessionMock{}

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Second*60)

	err := a.CreateUser("username", "hashedPassword")
	assert.Nil(t, err, "should be able to create a new use")

	err = a.DeleteUser("usernameOne")
	assert.NotNil(t, err, "delete user should return an error")
}

func TestAuthManager_DeleteUser_Cannot_Delete_All(t *testing.T) {
	tGenMock := &TokenGenMock{}
	idGenMock := &IdGenMock{}

	mockSession := &SessionMock{}
	mockSession.On("DeleteTokenContaining", "username").Return(nil)

	a := auth.NewAuthManager(mockSession, tGenMock, idGenMock, time.Second*60)

	err := a.CreateUser("username", "hashedPassword")
	assert.Nil(t, err, "should be able to create a new use")

	err = a.CreateUser("usernametwo", "hashedPassword")
	assert.Nil(t, err, "should be able to create a new use")

	err = a.DeleteUser("username")
	assert.Nil(t, err, "cannot delete last account")
}