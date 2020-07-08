package lockfree_queue

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type IntNode struct {
	item *int
	next *IntNode
}

func (node *IntNode) String() string {
	if node.item != nil {
		return fmt.Sprintf(`
	item:%-12v %-12p        next:%-12p        node:%-12p`,
			*node.item, node.item, node.next, node,
		)

	} else {
		return fmt.Sprintf(`
	item:%-12v %-12p        next:%-12p        node:%-12p`,
			node.item, node.item, node.next, node,
		)
	}
}

func (node *IntNode) casItem(cmp *int, val *int) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&node.item)), unsafe.Pointer(cmp), unsafe.Pointer(val))
}

func (node *IntNode) lazySetNext(val *IntNode) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&node.next)), unsafe.Pointer(val))
}

func (node *IntNode) casNext(cmp *IntNode, val *IntNode) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&node.next)), unsafe.Pointer(cmp), unsafe.Pointer(val))
}

type IntQueue struct {
	head  *IntNode
	tail  *IntNode
	state int32
}

func NewIntQueue() *IntQueue {
	null := &IntNode{item: nil, next: nil}
	return &IntQueue{head: null, tail: null}
}

func (queue *IntQueue) casHead(cmp *IntNode, val *IntNode) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&queue.head)), unsafe.Pointer(cmp), unsafe.Pointer(val))
}

func (queue *IntQueue) casTail(cmp *IntNode, val *IntNode) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&queue.tail)), unsafe.Pointer(cmp), unsafe.Pointer(val))
}

func (queue *IntQueue) updateHead(h *IntNode, p *IntNode) {
	if h != p && queue.casHead(h, p) {
		h.lazySetNext(h)
	}
}

func (queue *IntQueue) Offer(val *int) {
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

func (queue *IntQueue) Poll() (val *int) {
LabelRestartFromHead:
	for {
		var h, p, q *IntNode
		h = queue.head
		p = h
		for {
			var item = p.item

			if item != nil && p.casItem(item, nil) {
				if p != h {
					//queue.updateHead(h, ((q = p.next) != null) ? q: p);
					q = p.next
					if q != nil {
						queue.updateHead(h, q)
					} else {
						queue.updateHead(h, p)
					}
				}
				return item
			} else {
				q = p.next
				if q == nil {
					queue.updateHead(h, p)
					return nil
				} else if p == q {
					continue LabelRestartFromHead
				} else {
					p = q
				}
			}
		}
	}
}

func (queue *IntQueue) SyncPoll(valList *[]*int) (exit bool) {
	for {
		newVal := queue.Poll()
		if newVal == nil {
			if atomic.LoadInt32(&queue.state) == 1 {
				return true
			} else {
				return false
			}
		} else {
			*valList = append(*valList, newVal)
		}
	}
}

func (queue *IntQueue) Stop() {
	atomic.StoreInt32(&queue.state, 1)
}
