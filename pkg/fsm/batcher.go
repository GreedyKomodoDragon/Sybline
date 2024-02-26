package fsm

import (
	"sync"

	raft "github.com/GreedyKomodoDragon/raft"
	"github.com/vmihailenco/msgpack"
)

type Batcher interface {
	SendLog(*[]byte) (chan (*SyblineFSMResult), error)
}

type batcher struct {
	raftServer   raft.Raft
	inputChan    chan []byte
	currentIndex int
	dataSlice    [][]byte
	batchLength  int
	outputChans  []chan (*SyblineFSMResult)
	lock         *sync.Mutex
}

func NewBatcher(raftServer raft.Raft, batchLength int) Batcher {
	chans := make([]chan *SyblineFSMResult, batchLength)
	for i := 0; i < batchLength; i++ {
		chans[i] = make(chan *SyblineFSMResult, 1)
	}

	return &batcher{
		raftServer:   raftServer,
		inputChan:    make(chan []byte, batchLength),
		currentIndex: 0,
		dataSlice:    make([][]byte, batchLength),
		batchLength:  batchLength,
		outputChans:  chans,
		lock:         &sync.Mutex{},
	}
}

func (b *batcher) processLogs() error {

	for {
		data := <-b.inputChan
		b.dataSlice[b.currentIndex] = data
		b.currentIndex = (b.batchLength + 1) % b.batchLength

		if b.currentIndex != b.batchLength {
			continue
		}

		b.currentIndex = 0

		marshalledSlice, err := msgpack.Marshal(b.dataSlice)
		if err != nil {
			for i := 0; i < b.batchLength; i++ {
				b.outputChans[i] <- nil
			}
		}

		results, err := b.raftServer.ApplyLog(&marshalledSlice, raft.DATA_LOG)
		if err != nil {
			for i := 0; i < b.batchLength; i++ {
				b.outputChans[i] <- nil
			}
		}

		sybResult, ok := results.([]*SyblineFSMResult)
		if !ok {
			for i := 0; i < b.batchLength; i++ {
				b.outputChans[i] <- nil
			}
		}

		for i := 0; i < len(sybResult); i++ {
			b.outputChans[i] <- sybResult[i]
		}
	}
}

func (b *batcher) SendLog(data *[]byte) (chan (*SyblineFSMResult), error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	currentIndex := b.currentIndex
	b.inputChan <- *data

	return b.outputChans[currentIndex], nil
}
