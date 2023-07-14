package core

import (
	"errors"

	"github.com/hashicorp/go-hclog"
)

var ErrInvalidQueueName = errors.New("invalid queue name")

type ConsumerManager interface {
	GetMessages(name string, amount uint32, consumerID []byte, time int64) ([]Message, error)
	Ack(name string, id []byte, consumerID []byte) error
	BatchAck(name string, ids [][]byte, consumerID []byte) error
	Nack(name string, id []byte, consumerID []byte) error
	BatchNack(name string, ids [][]byte, consumerID []byte) error
}

func NewConsumerManager(queueManager QueueManager, logger hclog.Logger) ConsumerManager {
	return consumerManager{
		queueManager: queueManager,
		logger:       logger,
	}
}

type consumerManager struct {
	queueManager QueueManager
	logger       hclog.Logger
}

func (c consumerManager) GetMessages(name string, amount uint32, consumerID []byte, time int64) ([]Message, error) {
	if len(name) == 0 {
		return nil, ErrInvalidQueueName
	}

	messages, err := c.queueManager.GetMessage(name, amount, consumerID, time)
	if err != nil {
		c.logger.Error("failed to get message", name, err)
		return nil, err
	}

	return messages, nil
}

func (c consumerManager) Ack(name string, messageId []byte, consumerID []byte) error {
	if err := c.queueManager.Ack(name, consumerID, messageId); err != nil {
		c.logger.Error("failed to acknowledge message", "queueName", name, "id", messageId, "err", err)
		return err
	}

	return nil
}

func (c consumerManager) BatchAck(name string, messageIds [][]byte, consumerID []byte) error {
	if err := c.queueManager.BatchAck(name, consumerID, messageIds); err != nil {
		c.logger.Error("failed to batch acknowledge message", "queueName", name, "messageids", messageIds, "err", err)
		return err
	}

	return nil
}

func (c consumerManager) Nack(name string, id []byte, consumerID []byte) error {
	if err := c.queueManager.Nack(name, consumerID, id); err != nil {
		c.logger.Error("failed to nacknowledge message", "queueName", name, "id", id, "err", err)
		return err
	}

	return nil
}

func (c consumerManager) BatchNack(name string, messageIds [][]byte, consumerID []byte) error {
	if err := c.queueManager.BatchNack(name, consumerID, messageIds); err != nil {
		c.logger.Error("failed to batch nacknowledge messages", "queueName", name, "err", err)
		return err
	}

	return nil
}
