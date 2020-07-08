package lockfree_queue

import "sync"

type IntQueueSignal struct {
	IntQueue
	noticeGuard sync.Mutex
	notice      *sync.Cond
	exit        bool
}

func NewIntQueueSignal() *IntQueueSignal {
	null := &IntNode{item: nil, next: nil}

	ret := &IntQueueSignal{IntQueue: IntQueue{head: null, tail: null}}
	ret.notice = sync.NewCond(&ret.noticeGuard)
	return ret
}

func (queue *IntQueueSignal) Offer(val *int) {
	newNode := &IntNode{item: val, next: nil}
	var t, p, q *IntNode
	t = queue.tail
	p = t
	for {
		q = p.next
		if q == nil {
			if p.casNext(nil, newNode) {
				if p != t {
					queue.casTail(t, newNode)
				}
				queue.notice.Signal()
				return
			}
		} else if p == q {
			//p = (t != (t = tail)) ? t : head;
			if t != queue.tail {
				t = queue.tail
				p = t
			} else {
				t = queue.tail
				p = queue.head
			}
		} else {
			//p = (p != t && t != (t = tail)) ? t : q;
			if p != t && t != queue.tail {
				t = queue.tail
				p = t
			} else {
				t = queue.tail
				p = q
			}
		}
	}
}

func (queue *IntQueueSignal) SyncPoll(valList *[]*int) (exit bool) {
	for {
		for {
			newVal := queue.Poll()
			if newVal == nil {
				break
			} else {
				*valList = append(*valList, newVal)
			}
		}

		if len(*valList) == 0 {
			queue.noticeGuard.Lock()
			exit = queue.exit
			if exit {
				queue.noticeGuard.Unlock()
				return
			} else {
				queue.notice.Wait()
				queue.noticeGuard.Unlock()
			}
		} else {
			break
		}
	}
	return
}

func (queue *IntQueueSignal) Stop() {
	queue.noticeGuard.Lock()
	queue.exit = true
	queue.noticeGuard.Unlock()
	queue.notice.Signal()
}
