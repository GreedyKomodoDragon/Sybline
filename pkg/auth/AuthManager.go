package auth

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/GreedyKomodoDragon/raft"
	"golang.org/x/crypto/scrypt"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var ErrUserDoesNotExist = errors.New("user does not exist")
var ErrUsernameAlreadyTaken = errors.New("username already taken")
var ErrInvalidLogin = errors.New("invalid login details")
var ErrAtLeastOneAccount = errors.New("cannot delete all accounts, at least one must remain")

type Accounts struct {
	Accounts []string `json:"accounts"`
}

type AuthManager interface {
	CreateUser(username, password string) error
	DeleteUser(username string) error
	Login(username, password string) (string, error)
	LogOut(username, token string) error
	ChangePassword(username, oldPassword, newPassword string) (bool, error)
	GetConsumerIDViaMD(md *metadata.MD) ([]byte, error)
	GetConsumerID(username, token string) ([]byte, error)
	GetUsername(md *metadata.MD) (string, error)
	GetUsernameToken(md *metadata.MD) (string, string, error)
	UserExists(username string) bool
	GetAccounts() *Accounts
	Salt() string
}

func NewAuthManager(sessionHandler SessionHandler, tokenGen TokenGenerator, idGen IdGenerator, duration time.Duration, sessionServers []raft.Server, salt string) (AuthManager, error) {
	clients := []SessionClient{}
	for _, server := range sessionServers {
		conn, err := grpc.Dial(server.Address, server.Opts...)
		if err != nil {
			return nil, err
		}

		client := NewSessionClient(conn)
		clients = append(clients, client)

	}

	return &authManager{
		credentials:    []*saltPassword{},
		hashStrength:   32768,
		creationMux:    &sync.Mutex{},
		tokenName:      "syb-token",
		sessionHandler: sessionHandler,
		tokenGen:       tokenGen,
		idGen:          idGen,
		tokenDuration:  duration,
		follSessions:   NewFollowerSessions(clients),
		salt:           salt,
	}, nil
}

type saltPassword struct {
	username string
	salt     []byte
	hash     []byte
}

type authManager struct {
	credentials    []*saltPassword
	hashStrength   int
	creationMux    *sync.Mutex
	tokenName      string
	sessionHandler SessionHandler
	tokenGen       TokenGenerator
	idGen          IdGenerator
	tokenDuration  time.Duration
	follSessions   FollowerSessions
	salt           string
}

func (a *authManager) CreateUser(username, password string) error {
	a.creationMux.Lock()
	defer a.creationMux.Unlock()

	if user := a.getUserCredentials(username); user != nil {
		return ErrUsernameAlreadyTaken
	}

	buf := make([]byte, 128)
	if _, err := rand.Read(buf); err != nil {
		return err
	}

	hash, err := scrypt.Key([]byte(password), buf, a.hashStrength, 8, 1, 32)
	if err != nil {
		return err
	}

	a.credentials = append(a.credentials, &saltPassword{
		salt:     buf,
		hash:     hash,
		username: username,
	})

	return nil
}

func (a *authManager) Login(username, password string) (string, error) {
	saltPass := a.getUserCredentials(username)
	if saltPass == nil {
		return "", fmt.Errorf("failed to get saltPass")
	}

	hash, err := scrypt.Key([]byte(password), saltPass.salt, a.hashStrength, 8, 1, 32)
	if err != nil {
		return "", err
	}

	if !bytes.Equal(hash, saltPass.hash) {
		return "", ErrInvalidLogin
	}

	token := a.tokenGen.Generate()
	id := a.idGen.Generate()

	ttl := time.Now().Add(a.tokenDuration)

	if err := a.follSessions.PropagateSessions(token, username, id, ttl); err != nil {
		return "", err
	}

	if err := a.sessionHandler.AddToken(token, username, id, ttl); err != nil {
		return "", err
	}

	return token, nil
}

func (a *authManager) ChangePassword(username, oldPassword, newPassword string) (bool, error) {
	saltPass := a.getUserCredentials(username)
	if saltPass == nil {
		return false, ErrUserDoesNotExist
	}

	hash, err := scrypt.Key([]byte(oldPassword), saltPass.salt, a.hashStrength, 8, 1, 32)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(hash, saltPass.hash) {
		return false, nil
	}

	buf := make([]byte, 128)
	if _, err := rand.Read(buf); err != nil {
		return false, err
	}

	hash, err = scrypt.Key([]byte(newPassword), buf, a.hashStrength, 8, 1, 32)
	if err != nil {
		return false, err
	}

	a.deleteUsername(username)
	a.credentials = append(a.credentials, &saltPassword{
		salt:     buf,
		hash:     hash,
		username: username,
	})

	return true, nil
}

func (a *authManager) GetConsumerIDViaMD(md *metadata.MD) ([]byte, error) {
	username, token, err := a.GetUsernameToken(md)
	if err != nil {
		return nil, err
	}

	return a.sessionHandler.GetConsumerID(token, username)
}

func (a *authManager) GetConsumerID(username, token string) ([]byte, error) {
	return a.sessionHandler.GetConsumerID(token, username)
}

func (a *authManager) LogOut(username, token string) error {
	if err := a.follSessions.DeleteSessions(token, username); err != nil {
		return err
	}

	return a.sessionHandler.RemoveToken(token, username)
}

func (a *authManager) DeleteUser(username string) error {
	a.creationMux.Lock()
	defer a.creationMux.Unlock()

	if user := a.getUserCredentials(username); user == nil {
		return ErrUserDoesNotExist
	}

	if len(a.credentials) == 1 {
		return ErrAtLeastOneAccount
	}

	if err := a.sessionHandler.DeleteUser(username); err != nil {
		return err
	}

	a.deleteUsername(username)
	return nil
}

func (a *authManager) GetUsernameToken(md *metadata.MD) (string, string, error) {
	username, err := a.GetUsername(md)
	if err != nil {
		return "", "", err
	}

	tokenSlice := md.Get(a.tokenName)
	if len(tokenSlice) == 0 {
		return "", "", fmt.Errorf("missing token")
	}

	return username, tokenSlice[0], nil
}

func (a *authManager) getUserCredentials(username string) *saltPassword {
	for _, data := range a.credentials {
		if data.username == username {
			return data
		}
	}

	return nil
}

func (a *authManager) GetUsername(md *metadata.MD) (string, error) {
	userSlice := md.Get("username")
	if len(userSlice) == 0 {
		return "", fmt.Errorf("missing username")
	}

	return userSlice[0], nil
}

func (a *authManager) UserExists(username string) bool {
	for _, data := range a.credentials {
		if data.username == username {
			return true
		}
	}

	return false
}

func (a *authManager) GetAccounts() *Accounts {
	accounts := &Accounts{
		Accounts: []string{},
	}

	for _, data := range a.credentials {
		accounts.Accounts = append(accounts.Accounts, data.username)
	}

	return accounts
}

func (a *authManager) Salt() string {
	return a.salt
}

func (a *authManager) deleteUsername(username string) {
	for i, data := range a.credentials {
		if data.username == username {
			a.credentials[i] = a.credentials[len(a.credentials)-1] // Copy last element to index i.
			a.credentials[len(a.credentials)-1] = nil              // Erase last element (write zero value).
			a.credentials = a.credentials[:len(a.credentials)-1]   // Truncate slice.
			return
		}
	}
}
