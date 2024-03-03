package fsm

import (
	"sync"
	"time"

	raft "github.com/GreedyKomodoDragon/raft"
	"github.com/vmihailenco/msgpack"
)

// Batcher represents the interface for sending and processing logs in batches.
type Batcher interface {
	SendLog(*[]byte) (chan (*SyblineFSMResult), error)
	ProcessLogs()
}

// DataIndex represents the data structure holding the raw data and its identifier.
type DataIndex struct {
	data []byte
	id   int
}

// batcher is the concrete implementation of the Batcher interface.
type batcher struct {
	raftServer    raft.Raft
	inputChan     chan *DataIndex
	nextFreeIndex int
	dataSlice     [][]byte
	batchLength   int
	outputChans   *[]chan (*SyblineFSMResult)
	lock          *sync.Mutex
	wait          *sync.WaitGroup
	timer         *time.Timer
	batchLock     *sync.Mutex
}

// NewBatcher creates a new Batcher with the specified Raft server and batch length.
func NewBatcher(raftServer raft.Raft, batchLength int) Batcher {
	chans := make([]chan *SyblineFSMResult, batchLength)
	for i := 0; i < batchLength; i++ {
		chans[i] = make(chan *SyblineFSMResult, 1)
	}

	timer := time.NewTimer(500 * time.Millisecond)

	return &batcher{
		raftServer:    raftServer,
		inputChan:     make(chan *DataIndex, batchLength),
		nextFreeIndex: 0,
		dataSlice:     make([][]byte, batchLength),
		batchLength:   batchLength,
		outputChans:   &chans,
		lock:          &sync.Mutex{},
		wait:          &sync.WaitGroup{},
		timer:         timer,
		batchLock:     &sync.Mutex{},
	}
}

// ProcessLogs starts goroutines to process and send logs in batches.
func (b *batcher) ProcessLogs() {
	go func() {
		for {
			data := <-b.inputChan
			b.batchLock.Lock()
			b.dataSlice[data.id] = data.data
			if data.id != (b.batchLength - 1) {
				b.wait.Done()
				b.batchLock.Unlock()
				continue
			}

			b.nextFreeIndex = 0

			datas := b.dataSlice
			b.dataSlice = make([][]byte, b.batchLength)

			chans := make([]chan *SyblineFSMResult, b.batchLength)
			for i := 0; i < b.batchLength; i++ {
				chans[i] = make(chan *SyblineFSMResult, 1)
			}

			outputs := *b.outputChans
			b.outputChans = &chans

			b.sendLogs(&datas, &outputs)
			b.batchLock.Unlock()
			b.wait.Done()
		}
	}()

	go func() {
		for {
			<-b.timer.C
			b.batchLock.Lock()

			// If no items in the slice
			if b.dataSlice[0] == nil {
				b.batchLock.Unlock()
				b.timer.Reset(50 * time.Millisecond)
				continue
			}

			currentIndex := b.nextFreeIndex
			b.nextFreeIndex = 0

			datas := b.dataSlice[0:currentIndex]
			b.dataSlice = make([][]byte, b.batchLength)

			chans := make([]chan *SyblineFSMResult, b.batchLength)
			for i := 0; i < b.batchLength; i++ {
				chans[i] = make(chan *SyblineFSMResult, 1)
			}

			outputs := *b.outputChans
			b.outputChans = &chans

			b.sendLogs(&datas, &outputs)
			b.batchLock.Unlock()
			b.timer.Reset(50 * time.Millisecond)
		}
	}()
}

// sendLogs sends the batched logs to the Raft server and processes the results.
func (b *batcher) sendLogs(datas *[][]byte, outputs *[]chan *SyblineFSMResult) {
	marshalledSlice, err := msgpack.Marshal(*datas)
	if err != nil {
		for i := 0; i < len(*datas); i++ {
			(*outputs)[i] <- nil
		}
		return
	}

	results, err := b.raftServer.ApplyLog(&marshalledSlice, raft.DATA_LOG)
	if err != nil {
		for i := 0; i < len(*datas); i++ {
			(*outputs)[i] <- nil
		}
		return
	}

	sybResult, ok := results.([]*SyblineFSMResult)
	if !ok {
		for i := 0; i < len(*datas); i++ {
			(*outputs)[i] <- nil
		}
		return
	}

	for i := 0; i < len(sybResult); i++ {
		(*outputs)[i] <- sybResult[i]
	}
}

// SendLog sends a log entry to the batcher and returns the corresponding output channel.
func (b *batcher) SendLog(data *[]byte) (chan (*SyblineFSMResult), error) {
	indexData := &DataIndex{
		data: *data,
		id:   0,
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	currentIndex := b.nextFreeIndex
	indexData.id = currentIndex

	b.nextFreeIndex = (b.nextFreeIndex + 1) % b.batchLength

	b.wait.Add(1)
	b.inputChan <- indexData
	b.wait.Wait()

	return (*b.outputChans)[currentIndex], nil
}
