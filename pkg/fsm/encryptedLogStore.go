package fsm

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	raft "github.com/GreedyKomodoDragon/raft"

	"github.com/rs/zerolog/log"
)

type ecryptedLogStore struct {
	logs   *safeMap
	index  uint64
	term   uint64
	piping bool

	threshold  uint64
	currBatch  uint64
	persistMux *sync.Mutex
	key        []byte
	iv         []byte
}

func NewEcryptedLogStore(threshold uint64, key, iv []byte) (raft.LogStore, error) {
	// log directory - Create a folder/directory at a full qualified path
	if err := os.MkdirAll(raft.LOG_DIR, 0755); err != nil && !strings.Contains(err.Error(), "file exists") {
		return nil, err
	}

	return &ecryptedLogStore{
		logs: &safeMap{
			make(map[uint64]*raft.Log),
			sync.RWMutex{},
		},
		index:      1,
		term:       0,
		threshold:  threshold,
		persistMux: &sync.Mutex{},
		piping:     false,
		key:        key,
		iv:         iv,
	}, nil
}

func (l *ecryptedLogStore) AppendLog(log *raft.Log) error {
	if l.logs == nil {
		return fmt.Errorf("missing slice")
	}

	l.logs.Set(log.Index, log)

	go l.persistLog()
	return nil
}

func (l *ecryptedLogStore) SetLog(index uint64, log *raft.Log) error {
	if l.logs == nil {
		return fmt.Errorf("missing slice")
	}

	l.logs.Set(index, log)
	return nil
}

func (l *ecryptedLogStore) GetLog(index uint64) (*raft.Log, error) {
	if l.logs == nil {
		return nil, fmt.Errorf("missing slice")
	}

	log, ok := l.logs.Get(index)
	if ok {
		return log, nil
	}

	l.persistMux.Lock()
	defer l.persistMux.Unlock()

	// go to disk
	entries, err := os.ReadDir(raft.LOG_DIR)
	if err != nil {
		return nil, err
	}

	name := ""
	for _, entry := range entries {
		splitEntry := strings.Split(entry.Name(), raft.HYPEN)
		if len(splitEntry) < 2 {
			return nil, raft.ErrLogFormatWrong
		}

		l, err := strconv.ParseUint(splitEntry[0], 10, 64)
		if err != nil {
			return nil, err
		}

		h, err := strconv.ParseUint(splitEntry[1], 10, 64)
		if err != nil {
			return nil, err
		}

		if h > index && index >= l {
			name = raft.LOG_DIR + "/" + entry.Name()
			break
		}
	}

	if len(name) == 0 {
		return nil, fmt.Errorf("index missing: %v", index)
	}

	logs, err := l.readLogsFromDisk(name)
	if err != nil {
		return nil, err
	}

	for _, lg := range *logs {
		l.logs.Set(lg.Index, lg)
	}

	if log, ok := l.logs.Get(index); ok {
		return log, nil
	}

	return nil, fmt.Errorf("cannot find log: %v", index)
}

func (l *ecryptedLogStore) readLogsFromDisk(location string) (*[]*raft.Log, error) {
	logs := []*raft.Log{}
	file, err := os.Open(location)
	if err != nil {
		log.Error().Str("location", location).Err(err).Msg("unable to open file")
		return nil, err
	}

	defer file.Close()

	block, err := aes.NewCipher(l.key)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new cipher for reading in logs")
		return nil, err
	}

	// Create a stream cipher using the AES block and IV
	stream := cipher.NewCFBDecrypter(block, l.iv)
	reader := &cipher.StreamReader{S: stream, R: file}

	decoder := gob.NewDecoder(reader)
	if err := decoder.Decode(&logs); err != nil {
		log.Error().Str("location", location).Err(err).Msg("unable to decode file")
		return nil, err
	}

	return &logs, nil
}

func (l *ecryptedLogStore) UpdateCommited(index uint64) (bool, error) {
	if l.logs == nil {
		return false, fmt.Errorf("missing slice")
	}

	lg, ok := l.logs.Get(index)
	if !ok {
		return false, fmt.Errorf("cannot find log: %v", index)
	}

	lg.LeaderCommited = true

	// Also very brittle
	if lg.Index-1 != l.index && !l.piping {
		log.Info().Msg("missing a log, piping required")
		l.piping = true
		return true, nil
	}

	return false, nil
}

// writes the current batch of logs to disk
func (l *ecryptedLogStore) persistLog() error {
	l.persistMux.Lock()
	defer l.persistMux.Unlock()

	l.currBatch++

	// escape early if not ready to be push to cache
	if l.currBatch < l.threshold {
		return nil
	}

	l.currBatch = 0

	nextStart := uint64(0)

	entries, err := os.ReadDir(raft.LOG_DIR)
	if err != nil {
		log.Error().Err(err).Msg("unable to read snapshot file")
		return err
	}

	for _, entry := range entries {
		splitEntry := strings.Split(entry.Name(), raft.HYPEN)
		if len(splitEntry) < 2 {
			log.Error().Str("filename", entry.Name()).Msg("failed to split file")
			return fmt.Errorf("unable to split")
		}

		i, err := strconv.ParseUint(splitEntry[1], 10, 64)
		if err != nil {
			log.Error().Str("filename", splitEntry[1]).Err(err).Msg("failed to parse uint64")
			return err
		}

		if nextStart < i {
			nextStart = i
		}
	}

	// name of file
	nextStart += 1
	last := l.index - 5

	// happens when node is restarting
	if nextStart >= last {
		return nil
	}

	for k := nextStart; k <= last; k += l.threshold {
		lower := k
		upper := k + l.threshold
		if upper > last {
			upper = last
		} else {
			upper--
		}

		if upper-lower < 5 {
			break
		}

		logs := []*raft.Log{}
		for i := lower; i <= upper; i++ {
			lg, err := l.GetLog(i)
			if err != nil {
				log.Error().Uint64("index", i).Err(err).Msg("failed to get log when persisting")
				return err
			}

			logs = append(logs, lg)
		}

		fileLocation := fmt.Sprintf(raft.FILE_FORMAT, lower, upper)
		f, err := os.Create(fileLocation)
		if err != nil {
			log.Error().Str("fileLocation", fileLocation).Err(err).Msg("failed to create file")
			return err
		}

		// Create a new AES cipher block using a key and IV
		block, err := aes.NewCipher(l.key)
		if err != nil {
			log.Error().Err(err).Msg("failed to create new Cipher")
			return err
		}

		// Create a stream cipher using the AES block and IV
		stream := cipher.NewCFBEncrypter(block, l.iv)
		writer := &cipher.StreamWriter{S: stream, W: f}

		encoder := gob.NewEncoder(writer)
		if err := encoder.Encode(logs); err != nil {
			f.Close()
			// TODO: Delete file
			return err
		}

		f.Close()
		// delete logs after
		l.deleteRange(0, upper)
	}

	return nil
}

func (l *ecryptedLogStore) RestoreLogs(app raft.ApplicationApply) error {
	dir, err := os.ReadDir(raft.LOG_DIR)
	if err != nil {
		log.Error().Err(err).Msg("unable to read directory when restoring logs")
		return err
	}

	var entries raft.DirEntries
	for _, entry := range dir {
		entries = append(entries, entry)
	}

	sort.Sort(entries)

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			// Skip non-regular files
			continue
		}

		location := raft.LOG_DIR + "/" + entry.Name()
		logs, err := l.readLogsFromDisk(location)
		if err != nil {
			return err
		}

		for _, lg := range *logs {
			// update to latest term and index found in snapshot
			if lg.Term > l.term {
				l.term = lg.Term
			}

			if lg.Index > l.index {
				l.index = lg.Index
			}

			if _, err := app.Apply(*lg); err != nil {
				log.Error().Uint64("log", lg.Index).Err(err).Msg("unable to apply log")
			}
		}

	}

	return nil
}

func (l *ecryptedLogStore) ApplyFrom(index uint64, app raft.ApplicationApply) {
	for {
		log, ok := l.logs.Get(index)
		if !ok {
			l.index = index - 1
			l.piping = false
			return
		}

		// we ignore error as maybe intentional
		if log.LeaderCommited {
			app.Apply(*log)
		}

		index++
	}
}

func (l *ecryptedLogStore) IncrementIndex() {
	l.index++
}

func (l *ecryptedLogStore) IncrementTerm() {
	l.term++
}

func (l *ecryptedLogStore) GetLatestIndex() uint64 {
	return l.index
}

func (l *ecryptedLogStore) GetLatestTerm() uint64 {
	return l.term
}

func (l *ecryptedLogStore) deleteRange(start, finish uint64) {
	l.logs.DeleteRange(start, finish)
}

func (l *ecryptedLogStore) IsPiping() bool {
	return l.piping
}

func (l *ecryptedLogStore) SetPiping(isPiping bool) {
	l.piping = isPiping
}
