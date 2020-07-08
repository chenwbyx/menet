package lockfree_queue

import (
	"sync"
)

// 不限制大小，添加不发生阻塞，接收阻塞等待
type LockIntQueue struct {
	list      []*int
	listGuard sync.Mutex
	listCond  *sync.Cond
}

// 添加时不会发送阻塞
func (queue *LockIntQueue) Add(msg *int) {
	queue.listGuard.Lock()
	queue.list = append(queue.list, msg)
	queue.listGuard.Unlock()

	queue.listCond.Signal()
}

func (queue *LockIntQueue) Reset() {
	queue.list = queue.list[0:0]
}

// 如果没有数据，发生阻塞
func (queue *LockIntQueue) Pick(retList *[]*int) (exit bool) {

	queue.listGuard.Lock()

	for len(queue.list) == 0 {
		queue.listCond.Wait()
	}

	// 复制出队列

	for _, data := range queue.list {

		if data == nil {
			exit = true
			break
		} else {
			*retList = append(*retList, data)
		}
	}

	queue.Reset()
	queue.listGuard.Unlock()

	return
}
func (queue *LockIntQueue) Stop() {
	queue.Add(nil)
}

func NewLockIntQueue() *LockIntQueue {
	self := &LockIntQueue{}
	self.listCond = sync.NewCond(&self.listGuard)

	return self
}
