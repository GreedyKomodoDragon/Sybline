package fsm

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sybline/pkg/utils"
	"sync"

	"github.com/hashicorp/raft"
)

const (
	LOG_DIR     string = "./node_data/logs"
	HYPEN       string = "-"
	FILE_FORMAT string = "node_data/logs/%v-%v"
)

var (
	ErrNoLogs         = errors.New("there are no logs")
	ErrLogFormatWrong = errors.New("invalid log file parsed: not matching #-# format")
	ErrKeyNotFound    = errors.New("key not found")
)

// Error shown when log with specific index cannot be found
type ErrIndexMissing struct {
	Value uint64
}

func (e *ErrIndexMissing) Error() string {
	return fmt.Sprintf("cannot find log with index: %d", e.Value)
}

type record struct {
	key []byte
	val []byte
}

type SyblineStore struct {
	records   []*record
	logs      map[uint64]*raft.Log
	batchRate uint64
	currBatch uint64
	logMux    *sync.Mutex
	mapLock   *sync.RWMutex
}

func NewStableStore(batchRate uint64) *SyblineStore {
	return &SyblineStore{
		records: []*record{
			{
				key: []byte("CurrentTerm"),
				val: uint64ToBytes(0),
			},
		},
		logs:      map[uint64]*raft.Log{},
		batchRate: batchRate,
		currBatch: 0,
		logMux:    &sync.Mutex{},
		mapLock:   &sync.RWMutex{},
	}
}

func (s *SyblineStore) Set(key []byte, val []byte) error {
	s.records = append(s.records, &record{
		key: key,
		val: val,
	})

	return nil
}

func (s *SyblineStore) Get(key []byte) ([]byte, error) {
	for _, val := range s.records {
		if bytes.Equal(val.key, key) {
			return val.val, nil
		}
	}
	return nil, ErrKeyNotFound
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (s *SyblineStore) SetUint64(key []byte, val uint64) error {
	return s.Set(key, uint64ToBytes(val))
}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (s *SyblineStore) GetUint64(key []byte) (uint64, error) {
	val, err := s.Get(key)
	if err != nil {
		return 0, err
	}

	return bytesToUint64(val), nil
}

// FirstIndex returns the first index written. 0 for no entries.
func (s *SyblineStore) FirstIndex() (uint64, error) {
	if len(s.logs) == 0 {
		return 0, nil
	}

	return 1, nil
}

// LastIndex returns the last index written. 0 for no entries.
func (s *SyblineStore) LastIndex() (uint64, error) {
	if len(s.logs) == 0 {
		return 0, nil
	}

	var k uint64 = 0
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()

	for key := range s.logs {
		if key > k {
			k = key
		}
	}

	return k, nil
}

// GetLog gets a log entry at a given index.
func (s *SyblineStore) GetLog(index uint64, log *raft.Log) error {
	if len(s.logs) == 0 {
		return ErrNoLogs
	}

	s.mapLock.RLock()
	if lg, ok := s.logs[index]; ok {
		*log = *lg
		s.mapLock.RUnlock()
		return nil
	}
	s.mapLock.RUnlock()

	// if not in memory
	entries, err := os.ReadDir(LOG_DIR)
	if err != nil {
		return err
	}

	name := ""
	for _, entry := range entries {
		splitEntry := strings.Split(entry.Name(), HYPEN)
		if len(splitEntry) < 2 {
			return ErrLogFormatWrong
		}

		l, err := strconv.ParseUint(splitEntry[0], 10, 64)
		if err != nil {
			return err
		}

		h, err := strconv.ParseUint(splitEntry[1], 10, 64)
		if err != nil {
			return err
		}

		if h > index && index >= l {
			name = LOG_DIR + "/" + entry.Name()
			break
		}
	}

	if len(name) == 0 {
		return &ErrIndexMissing{Value: index}
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := s.restoreLogs(f, name); err != nil {
		return err
	}

	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	if lg, ok := s.logs[index]; ok {
		*log = *lg
		return nil
	}

	return &ErrIndexMissing{Value: index}
}

// StoreLog stores a log entry.
func (s *SyblineStore) StoreLog(log *raft.Log) error {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	s.logs[log.Index] = log
	return nil
}

// StoreLogs stores multiple log entries.
func (s *SyblineStore) StoreLogs(logs []*raft.Log) error {
	for _, log := range logs {
		s.StoreLog(log)
	}

	go s.persistLog(logs) // push to non-blocking thread
	return nil
}

// writes the current batch of logs to disk
func (s *SyblineStore) persistLog(logs []*raft.Log) error {
	s.logMux.Lock()
	defer s.logMux.Unlock()

	s.currBatch += uint64(len(logs))

	// escape early if not ready to be push to cache
	if s.currBatch < s.batchRate {
		return nil
	}

	s.currBatch = 0

	last, err := s.LastIndex()
	if err != nil {
		return err
	}

	nextStart := uint64(0)

	entries, err := os.ReadDir(LOG_DIR)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		splitEntry := strings.Split(entry.Name(), HYPEN)
		if len(splitEntry) < 2 {
			return ErrLogFormatWrong
		}

		i, err := strconv.ParseUint(splitEntry[1], 10, 64)
		if err != nil {
			return err
		}

		if nextStart < i {
			nextStart = i
		}
	}

	// name of file
	nextStart += 1
	last -= 1

	// happens when node is restarting
	if nextStart >= last {
		return nil
	}

	increments := incrementSlice(nextStart, last, s.batchRate)
	for k := 0; k < len(increments)-1; k++ {
		lower := increments[k]
		upper := increments[k+1]
		filename := fmt.Sprintf(FILE_FORMAT, lower, upper)

		f, err := os.Create(filename)
		if err != nil {
			return nil
		}

		defer f.Close()

		// perform a write
		log := &raft.Log{}

		for i := lower; i <= upper; i++ {
			if err := s.GetLog(i, log); err != nil {
				return err
			}

			logData, err := json.Marshal(log)
			if err != nil {
				return err
			}

			logData = append(logData, BREAK_SYMBOL...)
			if _, err := f.Write(logData); err != nil {
				return err
			}
		}

		// delete logs after
		s.DeleteRange(lower, upper)
		runtime.GC()
	}

	return nil
}

// DeleteRange deletes a range of log entries. The range is inclusive.
func (s *SyblineStore) DeleteRange(min, max uint64) error {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()

	for i := min; i <= max; i++ {
		delete(s.logs, i)
	}

	return nil
}

// Restore pulls in the snapshot file and attempts to rebuild logs and state via logs
func (s *SyblineStore) restoreLogs(rClose io.ReadCloser, name string) error {
	bts := []byte{}
	buffer := make([]byte, 1024)

	for {
		n, err := rClose.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		bts = append(bts, buffer...)

		// TODO: find a better way that does not require two traverisals
		indexes := utils.FindIndexes(bts, BREAK_SYMBOL)
		if len(indexes) == 0 {
			continue
		}

		previous := 0

		for i := 0; i < len(indexes); i++ {
			subSlice := bts[previous:indexes[i]]
			if len(subSlice) == 0 {
				continue
			}

			if err := s.extractStoreLog(&subSlice); err == nil {
				previous = indexes[i] + len(BREAK_SYMBOL)
			}
		}

		bts = bts[previous:]
	}

	return nil
}

// Takes bytes and then applies log to fsm if a command log
func (s *SyblineStore) extractStoreLog(by *[]byte) error {
	payload := raft.Log{}
	if err := json.Unmarshal(*by, &payload); err != nil {
		return err
	}

	return s.StoreLog(&payload)
}

// Converts a uint to a byte slice
func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

// Converts bytes to an integer
func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func incrementSlice(start, finish, increment uint64) []uint64 {
	var count uint64
	if finish < start {
		count = 0
	} else {
		count = (finish-start)/increment + 1
	}

	result := make([]uint64, count)
	var i uint64 = 0
	for i = 0; i < count-1; i++ {
		result[i] = start + i*increment
	}

	if count > 0 {
		result[count-1] = finish
	}

	return result
}
