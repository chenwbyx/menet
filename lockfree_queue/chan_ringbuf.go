package lockfree_queue

import (
	"menet/lockfree_queue/ring_buf_queue"
	"sync/atomic"
)

type ChanIntRingBuf struct {
	cacheQueue *ring_buf_queue.Queue
	queryQueue *ring_buf_queue.Queue
	syncChan   chan *int
	query      chan bool
	stop       chan bool
	exit       int32
}

func (queue *ChanIntRingBuf) Add(msg *int) {
	queue.syncChan <- msg
}

func (queue *ChanIntRingBuf) Reset() {
}

func (queue *ChanIntRingBuf) Pick(retList *[]*int) (exit bool) {
	if atomic.LoadInt32(&queue.exit) == 1 {
		return true
	}
	queue.query <- true
	<-queue.query
	for i := queue.queryQueue.Length(); i > 0; i-- {
		*retList = append(*retList, queue.queryQueue.Remove())
	}
	return
}

func (queue *ChanIntRingBuf) Stop() {
	queue.stop <- true
}

func (queue *ChanIntRingBuf) collect() {
	var state int8
	for {
		select {
		case data := <-queue.syncChan:
			queue.cacheQueue.Add(data)
		case <-queue.query:
			queue.cacheQueue, queue.queryQueue = queue.queryQueue, queue.cacheQueue
			switch state {
			case 0:
				queue.query <- true
			case 1, 2:
				queue.query <- true
				state += 1
			case 3:
				atomic.StoreInt32(&queue.exit, 1)
				queue.query <- true
				return
			}
		case <-queue.stop:
			state = 1
		}
	}
}

func NewChanIntRingBuf() *ChanIntRingBuf {
	self := &ChanIntRingBuf{
		syncChan:   make(chan *int),
		query:      make(chan bool),
		stop:       make(chan bool),
		queryQueue: ring_buf_queue.New(),
		cacheQueue: ring_buf_queue.New(),
	}
	go self.collect()
	return self
}
