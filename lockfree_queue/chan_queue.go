package lockfree_queue

import "sync/atomic"

type ChanIntQueue struct {
	cacheQueue []*int
	queryQueue []*int
	syncChan   chan *int
	queryBegin chan bool
	queryEnd   chan bool
	stop       chan bool
	exit       int32
}

func (queue *ChanIntQueue) Add(msg *int) {
	queue.syncChan <- msg
}

func (queue *ChanIntQueue) Reset() {
}

// 如果没有数据，发生阻塞
func (queue *ChanIntQueue) Pick(retList *[]*int) (exit bool) {
	if atomic.LoadInt32(&queue.exit) == 1 {
		return true
	}
	queue.queryBegin <- true
	<-queue.queryEnd
	*retList = append(*retList, queue.queryQueue...)
	queue.queryQueue = queue.queryQueue[:]
	return
}

func (queue *ChanIntQueue) Stop() {
	queue.stop <- true
}

func (queue *ChanIntQueue) collect() {
	var state int8
	for {
		select {
		case data := <-queue.syncChan:
			queue.cacheQueue = append(queue.cacheQueue, data)
		case <-queue.queryBegin:
			queue.cacheQueue, queue.queryQueue = queue.queryQueue, queue.cacheQueue
			switch state {
			case 0:
				queue.queryEnd <- true
			case 1, 2:
				queue.queryEnd <- true
				state += 1
			case 3:
				atomic.StoreInt32(&queue.exit, 1)
				queue.queryEnd <- true
				return
			}
		case <-queue.stop:
			state = 1
		}
	}
}

func NewChanIntQueue() *ChanIntQueue {
	self := &ChanIntQueue{
		syncChan:   make(chan *int),
		queryBegin: make(chan bool),
		queryEnd:   make(chan bool),
		stop:       make(chan bool),
	}
	go self.collect()
	return self
}
