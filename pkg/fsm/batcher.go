package fsm

import (
	"sync"

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
	raftServer   raft.Raft
	inputChan    chan *DataIndex
	currentIndex int
	dataSlice    [][]byte
	batchLength  int
	outputChans  []chan (*SyblineFSMResult)
	lock         *sync.Mutex
	wait         *sync.WaitGroup
}

func NewBatcher(raftServer raft.Raft, batchLength int) Batcher {
	chans := make([]chan *SyblineFSMResult, batchLength)
	for i := 0; i < batchLength; i++ {
		chans[i] = make(chan *SyblineFSMResult, 1)
	}

	return &batcher{
		raftServer:   raftServer,
		inputChan:    make(chan *DataIndex, batchLength),
		currentIndex: 0,
		dataSlice:    make([][]byte, batchLength),
		batchLength:  batchLength,
		outputChans:  chans,
		lock:         &sync.Mutex{},
		wait:         &sync.WaitGroup{},
	}
}

func (b *batcher) ProcessLogs() {

	// TODO: Need to add a timer if not enough logs pushed in
	for {
		data := <-b.inputChan
		b.dataSlice[data.id] = data.data
		if b.currentIndex != (b.batchLength - 1) {
			b.wait.Done()
			continue
		}

		b.currentIndex = 0

		datas := b.dataSlice
		b.dataSlice = make([][]byte, b.batchLength)
		b.wait.Done()

		marshalledSlice, err := msgpack.Marshal(datas)
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
	b.currentIndex = (b.batchLength + 1) % b.batchLength

	b.wait.Add(1)

	b.inputChan <- &DataIndex{
		data: *data,
		id:   currentIndex,
	}

	b.wait.Wait()

	return b.outputChans[currentIndex], nil
}
