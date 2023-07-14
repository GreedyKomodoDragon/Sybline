package fsm

import (
	"bytes"
	"io"

	"github.com/hashicorp/raft"
)

// snapshotSybline handles snapshot
type snapshotSybline struct {
	Data []byte
}

// Persist persist to disk. Return nil on success, otherwise return error.
func (s snapshotSybline) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()

	// Write the snapshot data to the sink
	if _, err := sink.Write(s.Data); err != nil {
		return err
	}

	return sink.Close()
}

// Release release the lock after persist snapshot.
// Release is invoked when we are finished with the snapshot.
func (s snapshotSybline) Release() {}

func (s snapshotSybline) Restore(source io.ReadCloser) error {
	defer source.Close()

	// Read the snapshot data from the source
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(source); err != nil {
		return err
	}

	// Assign the snapshot data to the snapshot object
	s.Data = buf.Bytes()
	return nil
}

// newSnapshotSybline is returned by an FSM in response to a snapshotNoop
// It must be safe to invoke FSMSnapshot methods with concurrent
// calls to Apply.
func newSnapshotSybline(data []byte) (raft.FSMSnapshot, error) {
	return &snapshotSybline{
		Data: data,
	}, nil
}
