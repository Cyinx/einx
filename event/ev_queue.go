package event

import (
	"github.com/Cyinx/einx/queue"
	"sync/atomic"
)

type EventChan chan bool
type EventQueue struct {
	ev_queue *queue.RWQueue
	ev_cond  EventChan

	w uint32
	n uint32
	q int32
}

func NewEventQueue() *EventQueue {
	q := &EventQueue{
		ev_queue: queue.NewRWQueue(),
		ev_cond:  make(EventChan, 128),
	}
	return q
}

func (q *EventQueue) SemaChan() EventChan {
	return q.ev_cond
}

func (q *EventQueue) Push(event EventMsg) {
	q.ev_queue.Push(event)
	atomic.AddInt32(&q.q, 1)

	if q.notify_one() == true {
		select {
		case q.ev_cond <- true:
		default:
		}
	}
}

func (q *EventQueue) Get(event_list []interface{}, count uint32) uint32 {
	if atomic.LoadInt32(&q.q) < 0 {
		return 0
	}
	read_count, _ := q.ev_queue.Get(event_list, count)
	atomic.AddInt32(&q.q, 0-int32(read_count))
	return read_count
}

func (q *EventQueue) notify_one() bool {
	for {
		n := atomic.LoadUint32(&q.n)
		if atomic.LoadUint32(&q.w) == n {
			return false
		}
		if atomic.CompareAndSwapUint32(&q.n, n, n+1) {
			return true
		}
	}
}

func (q *EventQueue) WaitNotify() bool {
	atomic.AddUint32(&q.w, 1)
	if atomic.LoadInt32(&q.q) > 0 {
		q.notify_one()
		return false
	}
	return true
}

func (q *EventQueue) WaiterWake() {
	q.notify_one()
}

func (q *EventQueue) count() int {
	return int(atomic.LoadInt32(&q.q))
}
