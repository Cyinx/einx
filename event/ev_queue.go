package event

import (
	"github.com/Cyinx/einx/queue"
	"sync/atomic"
)

type EventChan chan bool
type EventQueue struct {
	ev_queue *queue.RWQueue

	wait_count int32
	ev_cond    EventChan
}

func NewEventQueue() *EventQueue {

	queue := &EventQueue{
		ev_queue: queue.NewRWQueue(),
		ev_cond:  make(EventChan, 128),
	}
	atomic.AddInt32(&queue.wait_count, 1)
	return queue
}

func (this *EventQueue) GetChan() EventChan {
	return this.ev_cond
}

func (this *EventQueue) Push(event EventMsg) {
	this.ev_queue.Push(event)
	for {
		if atomic.CompareAndSwapInt32(&this.wait_count, 0, 0) == true {
			return
		}
		old_val := atomic.LoadInt32(&this.wait_count)

		if atomic.CompareAndSwapInt32(&this.wait_count, old_val, old_val-1) == true {
			break
		}
	}
	this.ev_cond <- true
}

func (this *EventQueue) Get(event_list []interface{}, count uint32) uint32 {
	read_count, left_count := this.ev_queue.Get(event_list, count)
	if left_count == 0 {
		atomic.AddInt32(&this.wait_count, 1)
	} else {
		this.ev_cond <- true
	}
	return read_count
}
