package lockfree_queue

type IntQueueChan struct {
	IntQueue
	notice chan bool
}

func NewIntQueueChan() *IntQueueChan {
	null := &IntNode{item: nil, next: nil}
	return &IntQueueChan{IntQueue: IntQueue{head: null, tail: null}, notice: make(chan bool)}
}

func (queue *IntQueueChan) Offer(val *int) {
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
				select {
				case queue.notice <- false:
				default:
				}
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

func (queue *IntQueueChan) SyncPoll(valList *[]*int) (exit bool) {
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
			exit = <-queue.notice
			if exit {
				return
			}
		} else {
			break
		}
	}
	return
}

func (queue *IntQueueChan) Stop() {
	queue.notice <- true
	//select {
	//case queue.notice <- true:
	//default:
	//}
}
