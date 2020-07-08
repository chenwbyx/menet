package lockfree_queue

import (
	"fmt"
	"sync"
	"testing"
)

func benchmarkQueue(push func(*int), pop func(*[]*int) bool, stop func()) {
	magicNum := 10
	goNum := 5000
	wg := &sync.WaitGroup{}
	pushFn := func() {
		for i := 0; i < magicNum; i++ {
			val := i
			push(&val)
		}
		wg.Done()
	}
	popFn := func() {
		var retList []*int
		var exist bool
		for {
			retList = retList[0:0]
			exist = pop(&retList)
			if exist {
				break
			}
		}
	}

	wg.Add(goNum)
	for i := 0; i < goNum; i++ {
		go pushFn()
	}
	wg.Wait()
	go stop()
	popFn()

}

func BenchmarkChanQueue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := NewChanIntQueue()
		benchmarkQueue(queue.Add, queue.Pick, queue.Stop)
	}
}

func BenchmarkChanRindBuf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := NewChanIntRingBuf()
		benchmarkQueue(queue.Add, queue.Pick, queue.Stop)
	}
}

func BenchmarkLockQueue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := NewLockIntQueue()
		benchmarkQueue(queue.Add, queue.Pick, queue.Stop)
	}
}

func BenchmarkLockRingBuf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := NewLockIntRingBuf()
		benchmarkQueue(queue.Add, queue.Pick, queue.Stop)
	}
}

func BenchmarkLockFreeQueue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := NewIntQueue()
		benchmarkQueue(queue.Offer, queue.SyncPoll, queue.Stop)
	}
}

func BenchmarkLockFreeQueueChan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := NewIntQueueChan()
		benchmarkQueue(queue.Offer, queue.SyncPoll, queue.Stop)
	}
}

func BenchmarkLockFreeQueueSignal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := NewIntQueueSignal()
		benchmarkQueue(queue.Offer, queue.SyncPoll, queue.Stop)
	}
}

/* 1r  5000w
BenchmarkChanQueue-4             	      50	  31443604 ns/op
BenchmarkChanRindBuf-4           	      50	  31563926 ns/op
BenchmarkLockQueue-4             	     100	  12493223 ns/op
BenchmarkLockRingBuf-4           	     100	  12292690 ns/op
BenchmarkLockFreeQueue-4         	     200	   9324797 ns/op
BenchmarkLockFreeQueueChan-4     	     200	   9074118 ns/op
BenchmarkLockFreeQueueSignal-4   	     200	   8993923 ns/op
*/

func TestNodeCasNext(t *testing.T) {

	var v1 int = 1
	var v2 int = 2

	a := &IntNode{item: &v1, next: nil}
	b := &IntNode{item: &v2, next: nil}

	node := &IntNode{item: nil, next: a}

	fmt.Println(node, a, b)
	fmt.Println(node.casNext(a, b))
	fmt.Println(node, a, b)

}

func TestNodeCasItem(t *testing.T) {

	var v1 int = 1
	var v2 int = 2

	a := &IntNode{item: &v1, next: nil}
	b := &IntNode{item: &v2, next: nil}

	node := &IntNode{item: &v1, next: nil}

	fmt.Println(node, a, b)
	fmt.Println(node.casItem(&v1, &v2))
	fmt.Println(node, a, b)

}

func TestQueue(t *testing.T) {
	wg := &sync.WaitGroup{}
	queue := NewIntQueue()
	wg.Add(2)

	go func() {
		for i := 0; i < 20; i++ {
			item := queue.Poll()
			if item != nil {
				fmt.Println(*item)
			} else {
				fmt.Println(nil)
			}

		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 10; i++ {
			var val = i
			queue.Offer(&val)
		}
		wg.Done()
	}()

	wg.Wait()
}
