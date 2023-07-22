package fsm

import (
	"errors"
	"time"

	"github.com/hashicorp/raft"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrCastRaftResponse = errors.New("unable to cast raft command result")
)

// CommandPayload is payload sent by system when calling raft.Apply(cmd []byte, timeout time.Duration)
type CommandPayload struct {
	Op   Operation `json:"o"`
	Data []byte    `json:"d"`
}

func SendRaftCommand(raftServer *raft.Raft, payloadType Operation, payload interface{}) (*ApplyResponse, error) {
	jsonBytes, err := msgpack.Marshal(payload)
	if err != nil {
		return nil, err
	}

	data, err := msgpack.Marshal(CommandPayload{
		Op:   payloadType,
		Data: jsonBytes,
	})

	if err != nil {
		return nil, err
	}

	applyFuture := raftServer.Apply(data, 4000*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return nil, err
	}

	res, ok := applyFuture.Response().(*ApplyResponse)
	if !ok {
		return nil, ErrCastRaftResponse
	}

	return res, res.Error
}
