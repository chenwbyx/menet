package mewnet

import (
	"sync"
)

const EMessageQueueDefaultCap = 4

// 不限制大小，添加不发生阻塞，接收阻塞等待
type MessageQueue struct {
	list      [][]byte
	listGuard sync.Mutex
	listCond  *sync.Cond
}

// 添加时不会发送阻塞
func (queue *MessageQueue) Add(msg []byte) {
	queue.listGuard.Lock()
	queue.list = append(queue.list, msg)
	queue.listGuard.Unlock()

	queue.listCond.Signal()
}

func (queue *MessageQueue) Reset() {
	queue.listGuard.Lock()
	queue.reset()
	queue.listGuard.Unlock()
}

func (queue *MessageQueue) reset() {
	queue.list = queue.list[0:0]
}

// 如果没有数据，发生阻塞
func (queue *MessageQueue) Pick(retList *[][]byte) (exit bool) {
	queue.listGuard.Lock()
	defer queue.listGuard.Unlock()

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

	queue.reset()
	return
}

func NewMessageQueue() *MessageQueue {
	return NewMessageQueueWithCap(EMessageQueueDefaultCap)
}

func NewMessageQueueWithCap(cap int) *MessageQueue {
	queue := &MessageQueue{}
	queue.listCond = sync.NewCond(&queue.listGuard)
	queue.list = make([][]byte, 0, cap)
	return queue
}
