package snowflake

import (
	"context"
	"sort"
	"sync/atomic"
)

type queue struct {
	ch   chan int64
	pos  int32
	head int64
	wake chan struct{}
}

func newQueue(bufferSize int, wake chan struct{}) *queue {
	return &queue{
		wake: wake,
		ch:   make(chan int64, bufferSize),
	}
}

func (q *queue) peek() (int64, bool) {
	if q.head <= 0 {
		if q.pos > 0 {
			q.head = <-q.ch
			atomic.AddInt32(&q.pos, -1)
		}
	}
	if v := q.head; v > 0 {
		return v, true
	}

	return 0, false
}

func (q *queue) remove() (int64, bool) {
	v, ok := q.peek()
	if !ok {
		return 0, ok
	}

	q.head = 0
	return v, true
}

func (q *queue) add(ctx context.Context, v int64) bool {
	select {
	case <-ctx.Done():
		return false
	case q.ch <- v:
		atomic.AddInt32(&q.pos, 1)
		go func() {
			select {
			case q.wake <- struct{}{}:
			default:
			}
		}()
		return true
	}
}

type queues []*queue

func (q queues) Len() int      { return len(q) }
func (q queues) Swap(i, j int) { q[j], q[i] = q[i], q[j] }
func (q queues) Less(i, j int) bool {
	v1, ok1 := q[i].peek()
	v2, ok2 := q[j].peek()
	if ok1 == ok2 {
		return v1 < v2
	}
	if ok1 {
		return true
	}
	return false
}

func (q queues) Next() (int64, bool) {
	sort.Sort(q)
	return q[0].remove()
}
