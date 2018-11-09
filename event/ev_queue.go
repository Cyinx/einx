package event

import (
	"github.com/Cyinx/einx/queue"
	"sync/atomic"
)

type EventChan chan bool
type EventQueue struct {
	ev_queue *queue.RWQueue

	wait_count  int32
	notifyCount int32
	ev_cond     EventChan
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
	atomic.AddInt32(&this.notifyCount, 1)
	if atomic.LoadInt32(&this.wait_count) > 0 {
		this.ev_cond <- true
	}
}

func (this *EventQueue) Get(event_list []interface{}, count uint32) uint32 {
	if atomic.LoadInt32(&this.notifyCount) < 0 {
		return 0
	}
	read_count, _ := this.ev_queue.Get(event_list, count)
	atomic.AddInt32(&this.notifyCount, int32(0-int32(read_count)))
	return read_count
}

func (this *EventQueue) WaiterWake() {
	atomic.AddInt32(&this.wait_count, 1)
}

func (this *EventQueue) NotifyCount() int {
	return int(atomic.LoadInt32(&this.notifyCount))
}
