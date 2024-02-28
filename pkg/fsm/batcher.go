package fsm

import (
	"sync"
	"time"

	raft "github.com/GreedyKomodoDragon/raft"
	"github.com/vmihailenco/msgpack"
)

type Batcher interface {
	SendLog(*[]byte) (chan (*SyblineFSMResult), error)
	ProcessLogs()
}

type DataIndex struct {
	data []byte
	id   int
}

type batcher struct {
	raftServer    raft.Raft
	inputChan     chan *DataIndex
	nextFreeIndex int
	dataSlice     [][]byte
	batchLength   int
	outputChans   []chan (*SyblineFSMResult)
	lock          *sync.Mutex
	wait          *sync.WaitGroup
	timer         *time.Timer
	batchLock     *sync.Mutex
}

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
		outputChans:   chans,
		lock:          &sync.Mutex{},
		wait:          &sync.WaitGroup{},
		timer:         timer,
		batchLock:     &sync.Mutex{},
	}
}

func (b *batcher) ProcessLogs() {
	go func() {
		for {
			data := <-b.inputChan
			b.batchLock.Lock()
			b.dataSlice[data.id] = data.data
			if b.nextFreeIndex != (b.batchLength - 1) {
				b.wait.Done()
				b.batchLock.Unlock()
				continue
			}

			currentIndex := b.nextFreeIndex
			b.nextFreeIndex = 0

			datas := b.dataSlice[0:currentIndex]
			b.dataSlice = make([][]byte, b.batchLength)
			b.wait.Done()
			b.batchLock.Unlock()

			b.sendLogs(&datas)
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
			b.batchLock.Unlock()

			b.sendLogs(&datas)
			b.timer.Reset(50 * time.Millisecond)
		}
	}()

}

func (b *batcher) sendLogs(datas *[][]byte) {
	marshalledSlice, err := msgpack.Marshal(*datas)
	if err != nil {
		for i := 0; i < b.batchLength; i++ {
			b.outputChans[i] <- nil
		}
		return
	}

	results, err := b.raftServer.ApplyLog(&marshalledSlice, raft.DATA_LOG)
	if err != nil {
		for i := 0; i < b.batchLength; i++ {
			b.outputChans[i] <- nil
		}
		return
	}

	sybResult, ok := results.([]*SyblineFSMResult)
	if !ok {
		for i := 0; i < b.batchLength; i++ {
			b.outputChans[i] <- nil
		}
		return
	}

	for i := 0; i < len(sybResult); i++ {
		b.outputChans[i] <- sybResult[i]
	}
}

func (b *batcher) SendLog(data *[]byte) (chan (*SyblineFSMResult), error) {
	indexData := &DataIndex{
		data: *data,
		id:   0,
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	currentIndex := b.nextFreeIndex
	b.nextFreeIndex = (b.nextFreeIndex + 1) % b.batchLength

	b.wait.Add(1)
	indexData.id = currentIndex

	b.inputChan <- indexData

	b.wait.Wait()

	return b.outputChans[currentIndex], nil
}
