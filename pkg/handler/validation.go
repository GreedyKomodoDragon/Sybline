package handler

import (
	"errors"
)

var (
	ErrInvalidData        = errors.New("data must container some information")
	ErrInvalidRoutingKey  = errors.New("routingKey cannot be empty")
	ErrInvalidQueueName   = errors.New("queue name cannot be empty")
	ErrInvalidAmount      = errors.New("must request more than zero")
	ErrInvalidSize        = errors.New("size must be more than zero")
	ErrInvalidId          = errors.New("id must not be empty")
	ErrInvalidUsername    = errors.New("username cannot be empty")
	ErrInvalidOldPassword = errors.New("old password cannot be empty")
	ErrInvalidNewPassword = errors.New("new password cannot be empty")
	ErrInvalidPassword    = errors.New("password cannot be empty")
)

func validateMessage(data []byte, routingKey string) error {
	if len(data) == 0 {
		return ErrInvalidData
	}

	if len(routingKey) == 0 {
		return ErrInvalidRoutingKey
	}

	return nil
}

func validateRequestMessages(queue string, amount uint32) error {
	if len(queue) == 0 {
		return ErrInvalidQueueName
	}

	if amount == 0 {
		return ErrInvalidAmount
	}

	return nil
}

func validateCreateQueue(routing, name string, size uint32) error {
	if len(routing) == 0 {
		return ErrInvalidRoutingKey
	}

	if len(name) == 0 {
		return ErrInvalidQueueName
	}

	if size == 0 {
		return ErrInvalidSize
	}

	return nil
}

func validateAck(queue string, id []byte) error {
	if len(queue) == 0 {
		return ErrInvalidQueueName
	}

	if len(id) == 0 {
		return ErrInvalidId
	}

	return nil
}

func validateBatchAck(queue string, ids [][]byte) error {
	if len(queue) == 0 {
		return ErrInvalidQueueName
	}

	if len(ids) == 0 {
		return ErrInvalidId
	}

	return nil
}

func validateChangePassword(username, oldPassword, newPassword string) error {
	if len(username) == 0 {
		return ErrInvalidUsername
	}

	if len(oldPassword) == 0 {
		return ErrInvalidOldPassword
	}

	if len(newPassword) == 0 {
		return ErrInvalidNewPassword
	}

	return nil
}

func validateUsernamePassword(username, newPassword string) error {
	if len(username) == 0 {
		return ErrInvalidUsername
	}

	if len(newPassword) == 0 {
		return ErrInvalidNewPassword
	}

	return nil
}

func validateRoutingKeyChange(queue, route string) error {
	if len(route) == 0 {
		return ErrInvalidRoutingKey
	}

	if len(queue) == 0 {
		return ErrInvalidQueueName
	}

	return nil
}
