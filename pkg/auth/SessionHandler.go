package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNoUserSessionExists = errors.New("unable to find user")

type SessionHandler interface {
	GetConsumerID(token string, username string) ([]byte, error)
	RemoveToken(token string, username string) error
	AddToken(token string, username string, consumerID []byte, ttl time.Duration) error
	DeleteTokenContaining(string) error
}

type timeSession struct {
	time time.Time
	id   []byte
}

func NewRedisSession(host string, database int, password string) (SessionHandler, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       database,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &redisSession{
		redisClient: redisClient,
		batchSize:   50,
		sessionIDs:  map[string]timeSession{},
	}, nil
}

type redisSession struct {
	redisClient *redis.Client
	sessionIDs  map[string]timeSession
	batchSize   int64
}

func (r *redisSession) GetConsumerID(token string, username string) ([]byte, error) {
	ctx := context.Background()

	key := token + "_" + username

	// use local session if possible
	sess, ok := r.sessionIDs[key]
	if ok {
		if t := time.Now(); sess.time.Before(t) {
			return sess.id, nil
		}

		delete(r.sessionIDs, key)
		return nil, ErrNoUserSessionExists
	}

	cmd := r.redisClient.Get(ctx, key)
	err := cmd.Err()
	if err == redis.Nil {
		return nil, ErrNoUserSessionExists
	}

	if err != nil {
		return nil, err
	}

	val, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	return []byte(val), nil
}

func (r *redisSession) RemoveToken(token string, username string) error {
	ctx := context.Background()
	return r.redisClient.Del(ctx, token+"_"+username).Err()
}

func (r *redisSession) AddToken(token string, username string, consumerID []byte, ttl time.Duration) error {
	ctx := context.Background()

	key := token + "_" + username

	r.sessionIDs[key] = timeSession{
		time: time.Now(),
		id:   consumerID,
	}

	cmd := r.redisClient.Set(ctx, token+"_"+username, consumerID, ttl)
	return cmd.Err()
}

func (r *redisSession) DeleteTokenContaining(subString string) error {
	ctx := context.Background()

	// Get total keys count
	totalKeys, err := r.redisClient.DBSize(ctx).Result()
	if err != nil {
		return err
	}

	// Calculate number of batches
	numBatches := (totalKeys + r.batchSize - 1) / r.batchSize

	for batch := int64(0); batch < numBatches; batch++ {
		// Scan keys in the current batch
		cursor := uint64(batch * r.batchSize)
		keys, nextCursor, err := r.redisClient.Scan(ctx, cursor, fmt.Sprintf("*%s*", subString), r.batchSize).Result()
		if err != nil {
			return err
		}

		// Delete keys containing the substring in a pipeline
		pipeline := r.redisClient.Pipeline()
		for _, key := range keys {
			if strings.Contains(key, subString) {
				pipeline.Del(ctx, key)
			}
		}

		if _, err := pipeline.Exec(ctx); err != nil {
			return err
		}

		// Break the loop if the cursor is 0, indicating the end of the iteration
		if nextCursor == 0 {
			break
		}
	}

	return nil
}
