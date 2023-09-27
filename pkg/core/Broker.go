package core

import (
	"errors"
	"sort"
	"sync"

	"github.com/rs/zerolog/log"
)

var ErrFailedQueueCreation = errors.New("failed to create new queue")
var ErrMissingRoutingKey = errors.New("missing routing key")
var ErrRoutingKeyDoesNotExist = errors.New("routing key does not exist")
var ErrQueueDoesNotExist = errors.New("queue name used does not exist")
var ErrQueueMustHaveOneRoutingKey = errors.New("each queue must have at least one routing key")
var ErrEmptyData = errors.New("cannot add message that is empty")

type Broker interface {
	CreateQueue(name, routingKey string, size, retryLimit uint32, hasDLQ bool) error
	DeleteQueue(name string) error
	AddMessage(routingKey string, data []byte, id []byte) error
	BatchAddMessage(routingKey string, data [][]byte, ids [][]byte) error
	AddRouteKey(routingKey, queueName string) error
	DeleteRoutingKey(routingKey, queueName string) error
}

func NewBroker(queueManager QueueManager) Broker {
	return broker{
		routing:      map[string][]string{},
		routingMut:   &sync.Mutex{},
		queueManager: queueManager,
	}
}

type broker struct {
	routing      map[string][]string
	routingMut   *sync.Mutex
	queueManager QueueManager
}

func (b broker) CreateQueue(name, routingKey string, size, retryLimit uint32, hasDLQ bool) error {
	if len(routingKey) == 0 {
		return ErrMissingRoutingKey
	}

	if err := b.queueManager.CreateQueue(name, size, retryLimit, hasDLQ); err != nil {
		log.Info().Str("queueName", name).Err(err).Msg("failed to create queue")
		return ErrFailedQueueCreation
	}

	b.routingMut.Lock()
	defer b.routingMut.Unlock()

	if _, ok := b.routing[routingKey]; ok {
		b.routing[routingKey] = append(b.routing[routingKey], name)

		// maintain increasing order -> tb used in linear locking
		sort.Strings(b.routing[routingKey])
		return nil
	}

	b.routing[routingKey] = []string{name}

	return nil
}

func (b broker) DeleteQueue(name string) error {
	b.routingMut.Lock()
	defer b.routingMut.Unlock()

	for key, value := range b.routing {
		for i, v := range value {
			if v == name {
				if err := b.queueManager.Delete(name); err != nil {
					return err
				}

				b.routing[key] = append(value[:i], value[i+1:]...)
				return nil
			}
		}
	}

	return nil
}

func (b broker) AddMessage(routingKey string, data []byte, id []byte) error {
	if len(data) == 0 {
		return ErrEmptyData
	}

	keys, ok := b.routing[routingKey]
	if !ok {
		log.Error().Str("routingKey", routingKey).Msg("failed to find routing key")
		return ErrRoutingKeyDoesNotExist
	}

	var wg sync.WaitGroup
	wg.Add(len(keys))

	for _, key := range keys {
		go func(key string) {
			defer wg.Done()
			// ignore any errors -> silent drop
			if err := b.queueManager.AddMessage(key, data, id); err != nil {
				log.Error().Bytes("id", id).Err(err).Str("key", key).Msg("failed to add message")
			}
		}(key)
	}

	wg.Wait()
	return nil
}

func (b broker) AddRouteKey(routingKey, queueName string) error {
	if len(routingKey) == 0 {
		return ErrMissingRoutingKey
	}

	if len(queueName) == 0 {
		return ErrInvalidQueueName
	}

	b.routingMut.Lock()
	defer b.routingMut.Unlock()

	if !b.queueManager.Exists(queueName) {
		return ErrQueueDoesNotExist
	}

	if _, ok := b.routing[routingKey]; ok {
		b.routing[routingKey] = append(b.routing[routingKey], queueName)

		// maintain increasing order -> tb used in linear locking
		sort.Strings(b.routing[routingKey])
		return nil
	}

	b.routing[routingKey] = []string{queueName}

	return nil
}

func (b broker) DeleteRoutingKey(routingKey, queueName string) error {
	if len(routingKey) == 0 {
		return ErrMissingRoutingKey
	}

	if len(queueName) == 0 {
		return ErrInvalidQueueName
	}

	b.routingMut.Lock()
	defer b.routingMut.Unlock()

	if !b.queueManager.Exists(queueName) {
		return ErrQueueDoesNotExist
	}

	if _, ok := b.routing[routingKey]; !ok {
		return ErrMissingRoutingKey
	}

	found := 0
	for _, value := range b.routing {
		if contains(value, queueName) {
			found++
			if found == 2 {
				break
			}
		}
	}

	if found == 1 {
		return ErrQueueMustHaveOneRoutingKey
	}

	idx := indexFunc(b.routing[routingKey], func(c string) bool { return c == queueName })

	b.routing[routingKey] = append(b.routing[routingKey][:idx], b.routing[routingKey][idx+1:]...)

	return nil
}

func (b broker) BatchAddMessage(routingKey string, datas [][]byte, ids [][]byte) error {
	// check all data exists
	for _, data := range datas {
		if len(data) == 0 {
			return ErrEmptyData
		}
	}

	keys, ok := b.routing[routingKey]
	if !ok {
		log.Error().Str("routingKey", routingKey).Msg("failed to find routing key")
		return ErrRoutingKeyDoesNotExist
	}

	var wg sync.WaitGroup
	wg.Add(len(keys))

	for i, key := range keys {
		go func(i int, key string) {
			defer wg.Done()

			// ignore any errors -> silent drop
			if err := b.queueManager.BatchAddMessage(key, datas, ids); err != nil {
				log.Error().Str("routingKey", routingKey).Msg("failed to add message")
			}
		}(i, key)
	}

	wg.Wait()

	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// IndexFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func indexFunc[E any](s []E, f func(E) bool) int {
	for i, v := range s {
		if f(v) {
			return i
		}
	}
	return -1
}
