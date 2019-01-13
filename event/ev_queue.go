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
	queue := &EventQueue{
		ev_queue: queue.NewRWQueue(),
		ev_cond:  make(EventChan, 128),
	}
	return queue
}

func (this *EventQueue) SemaChan() EventChan {
	return this.ev_cond
}

func (this *EventQueue) Push(event EventMsg) {
	this.ev_queue.Push(event)
	atomic.AddInt32(&this.q, 1)

	if this.notify_one() == true {
		select {
		case this.ev_cond <- true:
		default:
		}
	}
}

func (this *EventQueue) Get(event_list []interface{}, count uint32) uint32 {
	if atomic.LoadInt32(&this.q) < 0 {
		return 0
	}
	read_count, _ := this.ev_queue.Get(event_list, count)
	atomic.AddInt32(&this.q, 0-int32(read_count))
	return read_count
}

func (this *EventQueue) notify_one() bool {
	for {
		n := atomic.LoadUint32(&this.n)
		if atomic.LoadUint32(&this.w) == n {
			return false
		}
		if atomic.CompareAndSwapUint32(&this.n, n, n+1) {
			return true
		}
	}
}

func (this *EventQueue) WaitNotify() bool {
	atomic.AddUint32(&this.w, 1)
	if atomic.LoadInt32(&this.q) > 0 {
		this.notify_one()
		return false
	}
	return true
}

func (this *EventQueue) WaiterWake() {
	this.notify_one()
}

func (this *EventQueue) count() int {
	return int(atomic.LoadInt32(&this.q))
}
