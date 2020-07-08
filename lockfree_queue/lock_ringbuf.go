package lockfree_queue

import (
	"menet/lockfree_queue/ring_buf_queue"
	"sync"
)

// 不限制大小，添加不发生阻塞，接收阻塞等待
type LockIntRingBuf struct {
	list      *ring_buf_queue.Queue
	listGuard sync.Mutex
	listCond  *sync.Cond
}

// 添加时不会发送阻塞
func (queue *LockIntRingBuf) Add(msg *int) {
	queue.listGuard.Lock()
	//queue.list = append(queue.list, msg)
	queue.list.Add(msg)
	queue.listGuard.Unlock()

	queue.listCond.Signal()
}

func (queue *LockIntRingBuf) Reset() {
	//queue.list = queue.list[0:0]
}

// 如果没有数据，发生阻塞
func (queue *LockIntRingBuf) Pick(retList *[]*int) (exit bool) {

	queue.listGuard.Lock()

	//for len(queue.list) == 0 {
	for queue.list.Length() == 0 {
		queue.listCond.Wait()
	}

	// 复制出队列

	//for _, data := range queue.list {
	for i := queue.list.Length(); i > 0; i-- {
		data := queue.list.Remove()

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
func (queue *LockIntRingBuf) Stop() {
	queue.Add(nil)
}

func NewLockIntRingBuf() *LockIntRingBuf {
	self := &LockIntRingBuf{list: ring_buf_queue.New()}
	self.listCond = sync.NewCond(&self.listGuard)

	return self
}
