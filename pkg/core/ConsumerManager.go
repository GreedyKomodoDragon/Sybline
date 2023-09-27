package core

import (
	"errors"

	"github.com/rs/zerolog/log"
)

var ErrInvalidQueueName = errors.New("invalid queue name")

type ConsumerManager interface {
	GetMessages(name string, amount uint32, consumerID []byte, time int64) ([]Message, error)
	Ack(name string, id []byte, consumerID []byte) error
	BatchAck(name string, ids [][]byte, consumerID []byte) error
	Nack(name string, id []byte, consumerID []byte) error
	BatchNack(name string, ids [][]byte, consumerID []byte) error
}

func NewConsumerManager(queueManager QueueManager) ConsumerManager {
	return consumerManager{
		queueManager: queueManager,
	}
}

type consumerManager struct {
	queueManager QueueManager
}

func (c consumerManager) GetMessages(name string, amount uint32, consumerID []byte, time int64) ([]Message, error) {
	if len(name) == 0 {
		return nil, ErrInvalidQueueName
	}

	messages, err := c.queueManager.GetMessage(name, amount, consumerID, time)
	if err != nil {
		log.Error().Err(err).Str("name", name).Msg("failed to get message")
		return nil, err
	}

	return messages, nil
}

func (c consumerManager) Ack(name string, messageId []byte, consumerID []byte) error {
	if err := c.queueManager.Ack(name, consumerID, messageId); err != nil {
		log.Error().Err(err).Bytes("id", messageId).Str("queueName", name).Msg("failed to get message")
		return err
	}

	return nil
}

func (c consumerManager) BatchAck(name string, messageIds [][]byte, consumerID []byte) error {
	if err := c.queueManager.BatchAck(name, consumerID, messageIds); err != nil {
		log.Error().Err(err).Interface("messageids", messageIds).Str("queueName", name).Msg("failed to batch acknowledge message")
		return err
	}

	return nil
}

func (c consumerManager) Nack(name string, id []byte, consumerID []byte) error {
	if err := c.queueManager.Nack(name, consumerID, id); err != nil {
		log.Error().Err(err).Bytes("id", id).Str("queueName", name).Msg("failed to nacknowledge message")
		return err
	}

	return nil
}

func (c consumerManager) BatchNack(name string, messageIds [][]byte, consumerID []byte) error {
	if err := c.queueManager.BatchNack(name, consumerID, messageIds); err != nil {
		log.Error().Err(err).Interface("messageids", messageIds).Str("queueName", name).Msg("failed to batch nacknowledge message")
		return err
	}

	return nil
}
