package mewnet

import (
	"sync"
)

type WorkQueue struct {
	list      []func()
	listGuard sync.Mutex
}

func (queue *WorkQueue) Add(msg func()) {
	queue.listGuard.Lock()
	queue.list = append(queue.list, msg)
	queue.listGuard.Unlock()
}

func (queue *WorkQueue) Reset() {
	queue.listGuard.Lock()
	queue.reset()
	queue.listGuard.Unlock()
}

func (queue *WorkQueue) reset() {
	queue.list = make([]func(), 0)
}

func (queue *WorkQueue) Dump() (retList []func()) {
	queue.listGuard.Lock()
	retList = queue.list
	queue.reset()
	queue.listGuard.Unlock()
	return
}

func NewWorkQueue() *WorkQueue {
	return &WorkQueue{list: make([]func(), 0)}
}
