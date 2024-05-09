package fsm_test

import (
	"os"
	"sybline/pkg/fsm"
	"testing"
	"time"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/stretchr/testify/require"
)

// Test WriteLogs function
func TestWriteLogs(t *testing.T) {
	// Mock binaryStore instance
	binaryStore, err := fsm.NewBinaryLogStore(20)
	require.NoError(t, err)

	binStore, ok := binaryStore.(*fsm.BinaryStore)
	require.True(t, ok)

	for i := 1; i < 50; i++ {
		err := binStore.AppendLog(&raft.Log{
			Term:           1,
			Index:          uint64(i),
			LogType:        raft.DATA_LOG,
			Data:           []byte("test data"),
			LeaderCommited: true,
		})
		require.NoError(t, err)

		binStore.IncrementIndex()
	}

	time.Sleep(2 * time.Second)

	file, err := os.Open("./node_data/logs/1-20")
	require.NoError(t, err)

	defer file.Close()

	// Read the content of the file and assert if it's correct
	readLogs, err := binStore.ReadLogs(file)
	require.NoError(t, err)
	require.Equal(t, 20, len(readLogs))

	// Assert log content
	for i, log := range readLogs {
		expectedLog := raft.Log{
			Term:           1,
			Index:          uint64(i + 1),
			LogType:        raft.DATA_LOG,
			Data:           []byte("test data"),
			LeaderCommited: true,
		}

		require.Equal(t, expectedLog, log)
	}
}
